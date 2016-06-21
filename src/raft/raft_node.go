package raft

import (
	"time"	
	"sync"
	"logger"
	"fmt"
)


type Role int

const (
	Leader = 1
	Candidate = 2
	Follower = 3
)

const MinElectionTimeout = 150
const MaxElectionTimeout = 300

const HeartbeatTimeout = 50

const CurrentTermKey = "Current_Term"
const VotedForKey = "Voted_For_Key"

type trace struct {
	lastHeardFromLeader int64
	lastElectionSignal int64
}

type Node struct {
	config Config
	stable Stable
	transport Transport

	currentTerm uint64
	currentRole Role
	id string

	leader Peer

	votesGot uint32

	electionTimeout time.Duration
	heartbeatTimeout time.Duration

	electionTicker *time.Ticker
	heartbeatTicker *time.Ticker

	eventChannel chan interface{}

	quitChannel chan bool

	wg *sync.WaitGroup

	running bool

	handlers *handlers

	trace trace

	roleChange chan Role
}




func NewNode(id string,config Config,transport Transport,stable Stable) *Node {
	n := new(Node)
	n.id = id
	n.currentTerm = 0
	n.currentRole = Follower

	n.config = config
	n.stable = stable
	n.transport = transport

	n.setTermFromStable()

	n.setTimeoutValues()

	n.wg = &sync.WaitGroup{}

	n.eventChannel = make(chan interface{})
	n.quitChannel = make(chan bool)
	n.roleChange = make(chan Role)

	n.handlers = NewHandlers()
	
	return n
}


func (n *Node) Start() {

	n.startTimeSignals()
	n.loop()
	n.running = true
}

func (n *Node) loop() {

	logger.GetLogger().Log(fmt.Sprintf("%s - Starting node\n",n.id))
	go func() {
		defer n.wg.Done()
		n.wg.Add(1)

		for {
			select {
			case evt,ok := <- n.eventChannel :
				if ok {
					n.handlers.functions[n.currentRole](n,evt)
				}
			case quit,_ := <- n.quitChannel:
				if quit {
					logger.GetLogger().Log(fmt.Sprintf("%s - quitting\n",n.id))
					return
				}
			}
		}
	}()
}



func (n *Node) 	AppendEntry(entry Entry) (AppendResponse,error) {

	
	if entry.Term < n.currentTerm {
		return AppendResponse{ Reply:false,Term:n.currentTerm,From:n.id },nil
	}

	at := time.Now().UnixNano()

	logger.GetLogger().Log(fmt.Sprintf("%s - Got Append Entry from: %s at %d\n",n.id,entry.From,at))
	// heard from leader, set the trace
	n.trace.lastHeardFromLeader = at
	
	if entry.Term > n.currentTerm {
		n.higherTermDiscovered(entry.Term)
	}

	// do other checks here - according to raft paper
	
	return AppendResponse{ Reply: true, Term:n.currentTerm,From:n.id },nil
}

func (n *Node) RequestForVote(voteReq VoteRequest) (VoteResponse,error) {

	if voteReq.Term < n.currentTerm {
		return VoteResponse{ VoteGranted:false,From:n.id,TermToUpdate:n.currentTerm },nil
	}

	if voteReq.Term > n.currentTerm {
		n.higherTermDiscovered(voteReq.Term)
	}

	votedFor,ok := n.stable.Get(VotedForKey)
	if !ok || votedFor == "" || votedFor == voteReq.CandidateId {
		// other checks: !!!
		// perform log checks

		// store that the vote was granted
		n.stable.Store(VotedForKey,voteReq.CandidateId)
		// rasie votedFor event
		return VoteResponse{ VoteGranted:true,From:n.id,TermToUpdate:n.currentTerm },nil
		
	}
	
	return VoteResponse{ VoteGranted:false,From:n.id,TermToUpdate:n.currentTerm },nil
}

func (n *Node) CurrentRole() Role {
	return n.currentRole
}

func (n *Node) Stop() {
	n.running = false
	n.quitChannel <- true
	n.wg.Wait()
}

func (n *Node) IsRunning() bool {
	return n.running
}


func (n *Node) RoleChange() (chan Role) {
	return n.roleChange
}

func (n *Node) Id() string {
	return n.id
}
