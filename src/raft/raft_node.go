package raft

import (
	"time"	
	"sync"
	"logger"
	"fmt"
	"errors"
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

	appendResponseChannel chan AppendResponse
	voteResponseChannel chan VoteResponse
}




func NewNode(id string,config Config,transport Transport,stable Stable) *Node {
	n := new(Node)
	n.id = id
	n.currentTerm = 0
	

	n.config = config
	n.stable = stable
	n.transport = transport

	n.setTimeoutValues()

	n.wg = &sync.WaitGroup{}

	n.eventChannel = make(chan interface{})
	n.quitChannel = make(chan bool)
	n.roleChange = make(chan Role)
	n.appendResponseChannel = make(chan AppendResponse)
	n.voteResponseChannel = make(chan VoteResponse)

	n.handlers = NewHandlers()
	
	return n
}


func (n *Node) Start() {

	n.running = true
	n.currentRole = Follower
	n.setTermFromStable()

	n.startTimeSignals()
	n.loop()
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


/*
 Question:
Should I push AppendEntry and RequestForVote as events which have to dispacthed?
and later the below methods wait for a return value on different chennels??
*/

func (n *Node) 	AppendEntry(entry Entry) (AppendResponse,error) {

	
	if entry.Term < n.currentTerm {
		return AppendResponse{ Reply:false,Term:n.currentTerm,From:n.id },nil
	}	

	go func() {
		n.eventChannel <- &GotAppendEntryRequest{ entry: entry }
	}()

	// wait for response
	select {
	case responseGot, ok := <- n.appendResponseChannel:
		if ok {
			return responseGot,nil
		}
	}	
	return AppendResponse{ Reply: false, Term:n.currentTerm,From:n.id },errors.New(fmt.Sprintf("%s - Could not obtain a response for append entry",n.id))
	
}

func (n *Node) RequestForVote(voteReq VoteRequest) (VoteResponse,error) {

	if voteReq.Term < n.currentTerm {
		return VoteResponse{ VoteGranted:false,From:n.id,TermToUpdate:n.currentTerm },nil
	}

	
	go func() {
		n.eventChannel <- &GotRequestForVote{ voteRequest: voteReq }
	}()

	// wait for response
	select {
	case voteGot, ok := <- n.voteResponseChannel:
		if ok {
			return voteGot,nil
		}
	}
	
	return VoteResponse{ VoteGranted:false,From:n.id,TermToUpdate:n.currentTerm },nil
}

func (n *Node) CurrentRole() Role {
	return n.currentRole
}

func (n *Node) Stop() {
	n.running = false
	n.stopTimeSignals()
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
