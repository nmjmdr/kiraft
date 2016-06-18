package raft

func followerFn(n *Node,evt interface{}) {
	switch t := evt.(type) {
	case *ElectionNotice :
		//check when did this node hear from the leader?
		heard := n.haveHeardFromLeader(t.t)
		if !heard {
			// then have to transition to candidate and start election
			n.currentRole = Candidate
			n.currentTerm = n.currentTerm + 1
		} else {
			// ignore it
		}
		
	case *TimeForHeartbeat :
		// ignore this
		// the node is a follower, nothing to do
	}
}


