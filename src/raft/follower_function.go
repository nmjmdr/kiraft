package raft

import (
	"fmt"
	"logger"
	"time"
)

func followerFn(n *Node,evt interface{}) {
	switch t := evt.(type) {
	case *ElectionNotice :
		//check when did this node hear from the leader?
		
		newSignal := time.Now()
		heard := n.haveHeardFromLeader(newSignal)

		// reset the lastElectionSignal to new one
		n.trace.lastElectionSignal = newSignal.UnixNano()

		if !heard {
			// then have to transition to candidate and start election
			logger.GetLogger().Log(fmt.Sprintf("%s - have not heard from the leader, will send start election event\n",n.id))
			n.incrementTerm()
			n.setRole(Candidate)
			go func() {
				// send StartElection here
				n.eventChannel <- &StartElection{}
			}()
		} else {
			// ignore it
		}
		
	case *TimeForHeartbeat :
		// ignore this
		// the node is a follower, nothing to do
	case *GotVote:
		// ignore this, could have been a delayed vote response
	case *GotAppendEntryResponse:
		// could be a delayed response (this node could have been a leader earlier)

		// should we check explictily for reply flag in response?
		
		logger.GetLogger().Log(fmt.Sprintf("Node %s - recieved Append entry response from: %s while being a follower - ",n.id,t.response.From))
			
		if t.response.Term < n.currentTerm {
			logger.GetLogger().Log(fmt.Sprintf("With an older term, term: %d\n",t.response.Term))
			// should we do something here?
		} else {
			logger.GetLogger().Log(fmt.Sprintf("With a current or and a newer term, term: %d - ignoring it, the node has already transitioned to follower\n",t.response.Term))
		}
	/*
	// check if we need this? - as we are changing the way append entry is being handled
	case *HigherTermDiscovered:
		n.higherTermDiscovered(t.term)
*/
	case *GotAppendEntryRequest:
		n.handleAppendEntryRequest(t.entry)
	case *GotRequestForVote:
		n.handleRequestForVote(t.voteRequest)
	default :
		panic(fmt.Sprintf("%s - Unexpected event %T recieved by follower function\n",n.id,t))
	}


}


