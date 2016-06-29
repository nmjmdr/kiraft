package main

import (
	"strconv"
	"raft"
)


func makeNodes(numNodes int,config raft.Config,stable raft.Stable,transport *raft.InMemoryTransport) []raft.RaftNode  {

	nodes := make([]raft.RaftNode,numNodes)
	
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		nodes[i] = raft.NewNode(str,config,transport,stable)
		transport.SetNode(str,nodes[i])
	}
	return nodes
}

func getPeerConfiguration(numNodes int) raft.Config  {

	peers := make([]raft.Peer,numNodes)	
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		peers[i] =  raft.Peer{Id:str,Address:""}	
	}

	return raft.NewMockConfig(peers)
}
