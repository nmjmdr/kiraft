package raft

import (
	"time"
	"math/rand"
	"fmt"
	"logger"
	"strings"
	"strconv"
)

func (n *Node) handleRequestForVote(voteReq VoteRequest) {

	if voteReq.Term > n.currentTerm {
		n.higherTermDiscovered(voteReq.Term)
	}

	str,ok := n.stable.Get(VotedForKey)

	if !ok {
		// this is the fisrt time this node is voting
		n.stable.Store(VotedForKey,n.packVotedFor(voteReq.Term,voteReq.CandidateId))
		n.voteResponseChannel <- VoteResponse{ VoteGranted:true,From:n.id,TermToUpdate:n.currentTerm }
		return
	}

	votedTerm,votedFor := n.unpackVotedFor(str)

	// have we already voted for a different guy for request  term or a higher term??
	if  (votedFor != voteReq.CandidateId) && (votedTerm >= voteReq.Term) {		
		// we already voted for a different guy, deny the vote
		logger.GetLogger().Log(fmt.Sprintf("%s - has aldready voted for %s in term %d, hence denying vote request from %s with term: %d",n.id,votedFor,votedTerm,voteReq.CandidateId,voteReq.Term))
		n.voteResponseChannel <- VoteResponse{ VoteGranted:false,From:n.id,TermToUpdate:n.currentTerm }
		return
	}
	// other checks: !!!
	// perform log checks

	// store that the vote was granted
	n.stable.Store(VotedForKey,n.packVotedFor(voteReq.Term,voteReq.CandidateId))
	n.voteResponseChannel <- VoteResponse{ VoteGranted:true,From:n.id,TermToUpdate:n.currentTerm }		
	
}

func (n *Node) packVotedFor(term uint64,votedFor string) string {

	termString := strconv.FormatUint(term,10)
	return (termString + ":" + votedFor)
}

func (n *Node) unpackVotedFor(str string) (uint64,string) {

	splits := strings.Split(str,":")

	if splits == nil || len(splits) != 2 {
		panic(fmt.Sprintf("%s- Stable has an incorrect value for votedfor",n.id))
	}

	term,err := strconv.ParseUint(splits[0],10,64)
	if err != nil {
		panic(fmt.Sprintf("%s- Stable has an incorrect value for votedfor - %s",n.id,err))
	}

	return term,splits[1]
}


func (n *Node) handleAppendEntryRequest(entry Entry) {

	// first set the trace
	n.trace.lastHeardFromLeader = time.Now().UnixNano()

	// if a candidate and the term is atleast as much as the current term, step down
	if entry.Term >= n.currentTerm && n.currentRole == Candidate {
		// step down as candidate
		n.currentRole = Follower
		fmt.Printf("%s - got an append entry from %s, stepping down as candidate, becoming a follower\n",n.id,entry.From)
		logger.GetLogger().Log(fmt.Sprintf("%s - got an append entry from %s, stepping down as candidate, becoming a follower\n",n.id,entry.From))
	}

	
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
				fmt.Printf("%s node - RequestForVote got an error response from peer %s. Error detail: %s\n",n.id,p.Id,err)
				return
			}

			// generate got vote event
			n.eventChannel <- &GotVote{ response: voteResp }
		}(peer,voteReq)
	}
}


func (n *Node) higherTermDiscovered(term uint64) {
	n.setTerm(term)
	
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
	n.stable.Store(CurrentTermKey,n.currentTerm)
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
	n.stable.Store(CurrentTermKey,n.currentTerm)
	
}

func (n *Node) setTimeoutValues() {
	n.electionTimeout = getRandomTimeout(MinElectionTimeout,MaxElectionTimeout)
	n.heartbeatTimeout = time.Duration(HeartbeatTimeout * time.Millisecond)

	fmt.Printf("%s - election time out set to: %d\n,",n.id,n.electionTimeout)
	logger.GetLogger().Log(fmt.Sprintf("%s - election time out set to: %d\n,",n.id,n.electionTimeout))
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

