package raft

import (
	"time"
	"math/rand"
)


func (n *Node) haveHeardFromLeader(newElectionSignal time.Time) bool {


	heard :=  (newElectionSignal.UnixNano() >= n.trace.lastHeardFromLeader) &&  (n.trace.lastElectionSignal < n.trace.lastHeardFromLeader)		

	n.trace.lastElectionSignal = newElectionSignal.UnixNano()

	return heard
	
}

func (n *Node) setTimeoutValues() {
	n.electionTimeout = getRandomTimeout(MinElectionTimeout,MaxElectionTimeout)
	n.heartbeatTimeout = time.Duration(HeartbeatTimeout * time.Millisecond)
}

func getRandomTimeout(startRange int,endRange int) time.Duration {
	timeout := startRange + rand.Intn(endRange - startRange)

	return time.Duration(timeout) * time.Millisecond
}

func (n *Node) startTimeSignals() {
	n.electionTicker = time.NewTicker(n.electionTimeout)
	n.heartbeatTicker = time.NewTicker(n.heartbeatTimeout)

	go func() {
		for t := range n.electionTicker.C {
			n.eventChannel <- &ElectionNotice{t:t}
		}
	}()

	go func() {
		for t := range n.heartbeatTicker.C {
			n.eventChannel <- &TimeForHeartbeat{t:t}
		}
	}()
}

func (n *Node) setTermFromStable() {
	term,ok := n.stable.GetUint64(CurrentTermKey)
	if !ok {
		// probably this is the first time we are running this node
		n.currentTerm = 0
		n.stable.Store(CurrentTermKey,n.currentTerm)
		return
	}
	n.currentTerm = term
}

