package main

import (
	"testing"
	"raft"
)


func Test_StartStop(t *testing.T) {

	config := getPeerConfiguration(1)
	stable := raft.NewInMemoryStable()
	transport := raft.NewInMemoryTransport()
	
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
	stable := raft.NewInMemoryStable()
	transport := raft.NewInMemoryTransport()
	
	nodes := makeNodes(1,config,stable,transport)

	nodes[0].Start()
	nodes[0].Stop()
	
	_,err := transport.AppendEntry(raft.Entry{},raft.Peer{Id:"0"})
	if err == nil {
		t.Fatal("Transport should have returned error for AppendEntry")
	}

	_,err = transport.RequestForVote(raft.VoteRequest{},raft.Peer{Id:"0"})
	if err == nil {
		t.Fatal("Transport should have returned error for RequestForVote")
	}
	
}
