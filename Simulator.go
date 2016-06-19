package main

import (	
	"raft"
	"fmt"
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
	quitChannel chan bool
}

func NewSimulator(numNodes int) Simulator {
	s := new(sim)
	s.numNodes = numNodes

	s.config = getPeerConfiguration(s.numNodes)
	s.stable = raft.NewInMemoryStable()
	s.transport = raft.NewInMemoryTransport()
	
	s.nodes = makeNodes(s.numNodes,s.config,s.stable,s.transport)

	s.quitChannel = make(chan bool)

	return s
}

func (s *sim) Start() {

	for _,node := range s.nodes {
		// listen to role change
		go func(n raft.RaftNode) {
			for {
				select {
				case role,ok := <- n.RoleChange():
					if ok {
						fmt.Printf("%s - role: %d\n",n.Id(),role)
					}
				case <- s.quitChannel:
					return
				}
			}
		}(node)
		node.Start()
	}
}

func (s *sim) Stop() {

	
	for _,node := range s.nodes {
		node.Stop()
	}
	s.quitChannel <- true
}

