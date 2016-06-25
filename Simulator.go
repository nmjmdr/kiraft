package main

import (	
	"raft"
	"fmt"
	"errors"
)

type Simulator interface {
	Start()
	Stop()
	GetStates() []string
	StartNode(index int) error
	StopNode(index int) error
	IsRunning(index int) (bool,error)
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
				case _,ok := <- n.RoleChange():
					if ok {
						//fmt.Printf("%s - role: %d\n",n.Id(),role)
					}
				case <- s.quitChannel:
					return
				}
			}
		}(node)
		node.Start()
		fmt.Printf("%s - role: %d\n",node.Id(),node.CurrentRole())
	}
}

func (s *sim) GetStates() []string {

	arr := make([]string,0)

	for _,n := range s.nodes {
		arr = append(arr,fmt.Sprintf("%s (node) - %d",n.Id(),n.CurrentRole()))
		
	}
	return arr
}

func (s *sim) Stop() {

	
	for _,node := range s.nodes {
		node.Stop()
	}
	s.quitChannel <- true
}


func (s *sim) StartNode(index int) error {

	running,err := s.IsRunning(index)

	if err != nil {
		return err
	}

	if running {
		return errors.New("Node is already running")
	}
		
	s.nodes[index].Start()
	return nil
}

func (s *sim) StopNode(index int) error {

	running,err := s.IsRunning(index)

	if err != nil {
		return err
	}

	if !running {
		return errors.New("Node is already stopped")
	}

	s.nodes[index].Stop()
	return nil

}


func (s *sim) IsRunning(index int) (bool,error) {
	if index < 0 || index > len(s.nodes) {
		return false,errors.New("Wrong node index")
	}
	
	node := s.nodes[index]

	return node.IsRunning(),nil
}
