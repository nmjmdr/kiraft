package raft

import (
	"fmt"
	"time"
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
		n.askForVotes()	
	case *TimeForHeartbeat :
		// ignore this
		// the node is a candidate, nothing to do
	case *GotVote:
		handleGotVote(n,t)	
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

	if !t.response.VoteGranted {
		// vote was not granted, possibly the other node has already vote for another candidate
		return
	}

	n.votesGot++

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


