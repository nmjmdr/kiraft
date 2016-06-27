package raft

import (
	"fmt"
	"time"
	"logger"
)

func candidateFn(n *Node,evt interface{}) {
	switch t := evt.(type) {
	case *ElectionNotice :	
		// the node did not get elected,
		// increment the term, and restart the election
		//n.trace.lastElectionSignal = t.t.UnixNano()
		n.trace.lastElectionSignal = time.Now().UnixNano()
		n.incrementTerm()
		go func() {
			n.eventChannel <- &StartElection{}
		}()		
	case *StartElection:
		// reset votes Got
		n.votesGot = 0
		n.askForVotes()	
	case *TimeForHeartbeat :
		// ignore this
		// the node is a candidate, nothing to do
	case *GotVote:
		handleGotVote(n,t)
	/*
	// check if we need this?
	case *HigherTermDiscovered:
		n.higherTermDiscovered(t.term)
*/
	case *GotAppendEntryResponse:
		// could be a delayed response (this node could have been a leader earlier)

		// should we check explictily for reply flag in response?
		
		logger.GetLogger().Log(fmt.Sprintf("Node %s - recieved Append entry response from: %s while being a candidate - ",n.id,t.response.From))
			
		if t.response.Term < n.currentTerm {
			logger.GetLogger().Log(fmt.Sprintf("With an older term, term: %d\n",t.response.Term))
			// should we do something here?
		} else {
			logger.GetLogger().Log(fmt.Sprintf("With a current or and a newer term, term: %d - ignoring it, the node has already transitioned to follower and to a candidate\n",t.response.Term))
		}
	case *GotAppendEntryRequest:
		n.handleAppendEntryRequest(t.entry)
	case *GotRequestForVote:
		n.handleRequestForVote(t.voteRequest)
	default :
		panic(fmt.Sprintf("%s - Unexpected event %T recieved by candidate function\n",n.id,t))
	}
}


func handleGotVote(n *Node,t *GotVote) {
	// check for higher term
	if n.currentTerm < t.response.TermToUpdate {
		n.setRole(Follower)
		n.setTerm(t.response.TermToUpdate)
		return
	}

	fmt.Printf("%s - Got vote response - granted:%t, from: %s\n",n.id,t.response.VoteGranted,t.response.From)
	fmt.Printf("%s - current votes got: %d\n",n.id,n.votesGot)

	if !t.response.VoteGranted {
		// vote was not granted, possibly the other node has already vote for another candidate
		return
	}

	n.votesGot++

	fmt.Printf("%s - Votes Got after incrementing: %d\n",n.id,n.votesGot)

	majority := uint32(len(n.config.Peers())/2 + 1)
	if n.votesGot >= majority {
		// set as leader
		n.setRole(Leader)	
		go func() {
			// start the leader
			n.eventChannel <- &StartLeader{}
		}()
	}
}


