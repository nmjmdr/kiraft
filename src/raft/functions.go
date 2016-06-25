package raft

import (
	"time"
	"math/rand"
	"fmt"
	"logger"
)

func (n *Node) handleRequestForVote(voteReq VoteRequest) {

	if voteReq.Term > n.currentTerm {
		n.higherTermDiscovered(voteReq.Term)
	}

	votedFor,ok := n.stable.Get(VotedForKey)
	if !ok || votedFor == "" || votedFor == voteReq.CandidateId {
		// other checks: !!!
		// perform log checks

		// store that the vote was granted
		n.stable.Store(VotedForKey,voteReq.CandidateId)
		n.voteResponseChannel <- VoteResponse{ VoteGranted:true,From:n.id,TermToUpdate:n.currentTerm }
		
	} else {
		n.voteResponseChannel <- VoteResponse{ VoteGranted:false,From:n.id,TermToUpdate:n.currentTerm }
	}
}



func (n *Node) handleAppendEntryRequest(entry Entry) {

	// first set the trace
	n.trace.lastHeardFromLeader = time.Now().UnixNano()

	if entry.Term > n.currentTerm {
		n.higherTermDiscovered(entry.Term)
	}

	// do other checks here - according to raft paper
	// we will have to copy the log here

	go func() {
		n.appendResponseChannel <-  AppendResponse{ Reply:true, Term:n.currentTerm, From:n.id }
	}()
}

func (n *Node) haveHeardFromLeader(newElectionSignal time.Time) bool {


	heard :=  (newElectionSignal.UnixNano() >= n.trace.lastHeardFromLeader) &&  (n.trace.lastElectionSignal < n.trace.lastHeardFromLeader)

	logger.GetLogger().Log(fmt.Sprintf("%s - in have heard - last-heard-from-leader: %d, last-election-signal: %d, new-election-signal %d\n",n.id,n.trace.lastHeardFromLeader,n.trace.lastElectionSignal,newElectionSignal.UnixNano()))

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
		logger.GetLogger().Log(fmt.Sprintf("%s - discovered a higher term, will transition from %d to 3 (follower)\n",n.id,n.currentRole))
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

	logger.GetLogger().Log(fmt.Sprintf("%s - transitioned from %d to %d\n",n.id,n.currentRole,role))
	
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
		for _ = range n.electionTicker.C {
			n.eventChannel <- &ElectionNotice{}
		}
	}()

	go func() {
		for _ = range n.heartbeatTicker.C {			
			n.eventChannel <- &TimeForHeartbeat{}
			
		}
	}()
}


func (n *Node) stopTimeSignals() {
	n.electionTicker.Stop()
	n.heartbeatTicker.Stop()	
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

