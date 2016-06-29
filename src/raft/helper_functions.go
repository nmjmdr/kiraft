package raft

import (
	"strconv"
)


func makeNodes(numNodes int,config Config,transport *InMemoryTransport) []RaftNode  {

	nodes := make([]RaftNode,numNodes)
	
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		stable := NewInMemoryStable()
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
