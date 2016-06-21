package raft

import (
	"fmt"
	"logger"
)

func leaderFn(n *Node,evt interface{}) {
	switch t := evt.(type) {
	case *ElectionNotice :
		// ignore this, the node is the leader
	case *TimeForHeartbeat :
		// send heartbeat
		logger.GetLogger().Log(fmt.Sprintf("%s - Sending heartbeat\n",n.id))
		n.sendHeartbeat()
	case *GotVote:
		// ignore this, might be a delayed vote
	case *StartLeader:
		// send heartbeats to all peers
		logger.GetLogger().Log(fmt.Sprintf("%s - Sending heartbeat\n",n.id))
		n.sendHeartbeat()
	case *GotAppendEntryResponse:
		
		if t.response.Term > n.currentTerm {
			n.higherTermDiscovered(t.response.Term)
			return
		}

		// do other things - like maintaining the pointer to all followers
		// for log later here
	
	default :
	panic(fmt.Sprintf("Unexpected event %T recieved by leader function\n",t))
	}
}


