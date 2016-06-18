package main

import (	
	"raft"
)

type Simulator interface {
	Start()
	Stop()
}

type sim struct {
	nodes []raft.RaftNode
	config raft.Config
	stable raft.Stable
	transport *raft.InMemoryTransport
	numNodes int
}

func NewSimulator(numNodes int) Simulator {
	s := new(sim)
	s.numNodes = numNodes

	s.config = getPeerConfiguration(s.numNodes)
	s.stable = raft.NewInMemoryStable()
	s.transport = raft.NewInMemoryTransport()
	
	s.nodes = makeNodes(s.numNodes,s.config,s.stable,s.transport)	

	return s
}

func (s *sim) Start() {

	for _,node := range s.nodes {
		node.Start()
	}
}

func (s *sim) Stop() {

	for _,node := range s.nodes {
		node.Stop()
	}
}

