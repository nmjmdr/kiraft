package raft

import (
	"testing"
	"strconv"
	"time"
)


func makeNodes(numNodes int,config Config,stable Stable,transport *InMemoryTransport) []RaftNode  {

	nodes := make([]RaftNode,numNodes)
	
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		nodes[i] = NewNode(str,config,transport,stable)
		transport.SetNode(str,nodes[i])
	}
	return nodes
}

func getPeerConfiguration(numNodes int) Config  {

	peers := make([]Peer,numNodes)	
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		peers[i] =  Peer{Id:str,Address:""}	
	}

	return NewMockConfig(peers)
}

func waitForStateChanges(n RaftNode,nChanges int) []Role {

	states := make([]Role,nChanges)
	
	for i:=0;i<nChanges;i++ {
		select {
		case role,ok := <- n.RoleChange():
			if ok {
				states[i] = role
			}
		}		
	}
	return states
}


func Test_StartStop(t *testing.T) {

	config := getPeerConfiguration(1)
	stable := NewInMemoryStable()
	transport := NewInMemoryTransport()
	
	nodes := makeNodes(1,config,stable,transport)

	for _,n := range nodes {		
		n.Start()
	}

	for _,n := range nodes {
		n.Stop()
	}
	for _,n := range nodes {
		if(n.IsRunning()) {
			t.Fatal("Node is still running")
		}
	}
}


func Test_Stop_Transport_Err(t *testing.T) {

	config := getPeerConfiguration(1)
	stable := NewInMemoryStable()
	transport := NewInMemoryTransport()
	
	nodes := makeNodes(1,config,stable,transport)

	nodes[0].Start()
	nodes[0].Stop()
	
	_,err := transport.AppendEntry(Entry{},Peer{Id:"0"})
	if err == nil {
		t.Fatal("Transport should have returned error for AppendEntry")
	}

	_,err = transport.RequestForVote(VoteRequest{},Peer{Id:"0"})
	if err == nil {
		t.Fatal("Transport should have returned error for RequestForVote")
	}
	
}


func Test_SingleNode(t *testing.T) {

	config := getPeerConfiguration(1)
	stable := NewInMemoryStable()
	transport := NewInMemoryTransport()
	
	nodes := makeNodes(1,config,stable,transport)

	// reduce the time out for unit test
	n,_ := nodes[0].(*Node)

	n.setHeartbeatTimeout(1 * time.Millisecond)
	n.setElectionTimeout(10 * time.Millisecond)

	nodes[0].Start()

	states := waitForStateChanges(nodes[0],2)
	

	nodes[0].Stop()

	// now verify the states
	expectedStates := []Role{ Candidate, Leader }

	for i:=0;i<len(states);i++ {
		if states[i] != expectedStates[i] {
			t.Fatal("The node is not in the expected state")
		}
	}
}
