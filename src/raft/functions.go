package raft

import (
	"time"
	"math/rand"
	"fmt"
	"logger"
)


func (n *Node) haveHeardFromLeader(newElectionSignal time.Time) bool {


	heard :=  (newElectionSignal.UnixNano() >= n.trace.lastHeardFromLeader) &&  (n.trace.lastElectionSignal < n.trace.lastHeardFromLeader)

	return heard
	
}

func (n *Node) askForVotes() {

	//voteReq := VoteRequest{ Term:n.currentTerm,CandidateId:n.id,LastLogIndex:0,LastLogTerm:0 }

	logger.GetLogger().Log(fmt.Sprintf("%s - asking for votes\n",n.id))

	// self vote
	go func(n *Node) {
		gotVote := GotVote{ response : VoteResponse { VoteGranted:true, From:n.id,TermToUpdate:0 } }

		n.eventChannel <- &gotVote
	}(n)
	
	peers := n.config.Peers()
	voteReq := VoteRequest { Term : n. currentTerm, CandidateId: n.id, LastLogIndex:0, LastLogTerm:0 }
	for _,peer := range peers {

		if peer.Id == n.id {
			continue
		}
		
		go func(p Peer,vreq VoteRequest) {		
			voteResp,err := n.transport.RequestForVote(vreq,p)
			if err != nil {
				logger.GetLogger().Log(fmt.Sprintf("%s node - RequestForVote got an error response from peer %s. Error detail: %s\n",n.id,p.Id,err))
				return
			}

			// generate got vote event
			n.eventChannel <- &GotVote{ response: voteResp }
		}(peer,voteReq)
	}
}


func (n *Node) higherTermDiscovered(term uint64) {
	n.currentTerm = term

	if n.currentRole != Follower {
		n.setRole(Follower)
	}
}

func (n *Node) sendHeartbeat() {

	peers := n.config.Peers()
	entry := Entry{ Term : n.currentTerm, From: n.id }
	for _,peer := range peers {

		if peer.Id == n.id {
			continue
		}
		
		go func(p Peer,e Entry) {		
			appendResponse,err := n.transport.AppendEntry(e,p)
			if err != nil {
				logger.GetLogger().Log(fmt.Sprintf("%s node - Append Entry got an error response from peer %s. Error detail: %s\n",n.id,p.Id,err))
				return
			}

			// generate append entry response
			n.eventChannel <- &GotAppendEntryResponse{ response: appendResponse }
		}(peer,entry)
	}
}

func (n *Node) setHeartbeatTimeout(d time.Duration) {
	n.heartbeatTimeout = d
}

func (n *Node) setElectionTimeout(d time.Duration) {
	n.electionTimeout = d
}

func (n *Node) incrementTerm() {
	n.currentTerm = n.currentTerm + 1
}

func (n *Node) setRole(role Role) {
	n.currentRole = role
	go func(n *Node,role Role) {
		n.roleChange <- role
	}(n,role)
}

func (n *Node) setTerm(term uint64) {
	n.currentTerm = term
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

