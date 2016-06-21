package raft

import (
	"fmt"
	"logger"
)

func followerFn(n *Node,evt interface{}) {
	switch t := evt.(type) {
	case *ElectionNotice :
		//check when did this node hear from the leader?
		heard := n.haveHeardFromLeader(t.t)

		// reset the lastElectionSignal to new one
		n.trace.lastElectionSignal = t.t.UnixNano()

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
	default :
	panic(fmt.Sprintf("Unexpected event %T recieved by follower function\n",t))
	}


}


