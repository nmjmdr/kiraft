package raft

import (
	"fmt"
	"errors"
)

type InMemoryTransport struct {
	m map[string]RaftNode
}

func NewInMemoryTransport() *InMemoryTransport {
	i :=new(InMemoryTransport)
	i.m = make(map[string]RaftNode)
	return i
}


func (i *InMemoryTransport) SetNode(id string,n RaftNode) {
	i.m[id] = n
}


func (i *InMemoryTransport) AppendEntry(entry Entry,peer Peer) (AppendResponse,error) {

	node,ok := i.m[peer.Id]

	if !ok {
		panic(fmt.Sprintf("%s peer not found in transport map\n",peer.Id))
	}

	if !node.IsRunning() {
		return AppendResponse{},errors.New(fmt.Sprintf("%s peer is not running",peer.Id))
	}


	return node.AppendEntry(entry)
	
	
	
}


func (i *InMemoryTransport) RequestForVote(vreq VoteRequest,peer Peer) (VoteResponse,error) {
	node,ok := i.m[peer.Id]

	if !ok {
		panic(fmt.Sprintf("%s peer not found in transport map\n",peer.Id))
	}
	
	if !node.IsRunning() {
		fmt.Printf("%s - is not running, returning error\n",node.Id())
		return VoteResponse{},errors.New(fmt.Sprintf("%s peer is not running",peer.Id))
	}
	
	return node.RequestForVote(vreq)
	
}
