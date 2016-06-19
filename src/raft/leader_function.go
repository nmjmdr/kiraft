package raft

import (
	"fmt"
)

func leaderFn(n *Node,evt interface{}) {
	switch t := evt.(type) {
	case *ElectionNotice :
		// ignore this, the node is the leader
	case *TimeForHeartbeat :
		// send heartbeat
	case *GotVote:
		// ignore this, might be a delayed vote
	case *StartLeader:
		// send heartbeats to all peers
		n.sendHeartbeat()
	case *GotAppendEntryResponse:
		// handle discovery of higher term here
	
	default :
	panic(fmt.Sprintf("Unexpected event %T recieved by leader function\n",t))
	}
}


