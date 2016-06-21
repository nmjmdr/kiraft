package raft

import (
	"testing"
	"strconv"
	"time"
	"sync"
	"fmt"
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



func waitForLeaderRole(nodes [] RaftNode) RaftNode  {

	gotLeader := make(chan bool)
	
	wg := &sync.WaitGroup{}
	var leader RaftNode

	
	wg.Add(1)
	
	for _,node := range nodes {

		go func(node RaftNode) {

			for {
				select {
				case role,ok := <- node.RoleChange():
					if ok {						
						if role == Leader {
							leader = node
							gotLeader <- true
							wg.Done()
							return
						}
					}
				case <- gotLeader:
					return				
				}
			}
		}(node)	
	}

	wg.Wait()

	return leader
}


func getALeader(numNodes int) ([]RaftNode,RaftNode) {
	
	
	config := getPeerConfiguration(numNodes)
	stable := NewInMemoryStable()
	transport := NewInMemoryTransport()
	
	nodes := makeNodes(numNodes,config,stable,transport)


	
	// reduce the time out for unit test
	for _,node := range nodes {
		n,_ := node.(*Node)
		n.setHeartbeatTimeout(10)
		n.setElectionTimeout(getRandomTimeout(50,100))
	}


	for _,node := range nodes {
		node.Start()
	}


	leader := waitForLeaderRole(nodes)

	return nodes,leader

}

func Test_MultipleNode(t *testing.T) {

	numNodes := 3
	nodes,leader := getALeader(numNodes)
	
	t.Log(fmt.Sprintf("Leader: %s\n",leader.Id()))
	

	checkLeaderAmongN(nodes,t)

}

func checkLeaderAmongN(nodes []RaftNode,t *testing.T) {
// there should be one leader and two followers
	followerCount := 0
	leaderCount := 0
	for _,node := range nodes {
		if node.CurrentRole() == Leader {
			leaderCount++
		} else if node.CurrentRole() == Follower {
			followerCount++
		} else {
			t.Fatal(fmt.Sprintf("%s is in unexpected state: %d",node.Id(),node.CurrentRole()))
		}
	}

	if leaderCount != 1 || followerCount != len(nodes) - 1 {
		t.Fatal("The nodes are in unexpected states")
	}
}


func Test_Stop_Leader(t *testing.T) {

	numNodes := 3
	nodes,leader := getALeader(numNodes)
	
	t.Log(fmt.Sprintf("Leader: %s\n",leader.Id()))

	checkLeaderAmongN(nodes,t)

	// now we will stop the leader
	wg := &sync.WaitGroup{}
	wg.Add(1)
	var newLeader RaftNode
	go func(w *sync.WaitGroup) {
		newLeader = waitForLeaderRole(nodes)
		w.Done()
	}(wg)

	leader.Stop()
	wg.Wait()

	t.Log(fmt.Sprintf("New leader - %s\n",newLeader.Id()))

	newNodes := make([]RaftNode,0)

	for _,node := range nodes {
		if node.IsRunning() {
			newNodes = append(newNodes,node)
		}
	}

	checkLeaderAmongN(newNodes,t)
}



func Test_Stop_Leader_Then_Start(t *testing.T) {

	numNodes := 3
	nodes,leader := getALeader(numNodes)
	
	t.Log(fmt.Sprintf("Leader: %s\n",leader.Id()))

	checkLeaderAmongN(nodes,t)

	// now we will stop the leader
	wg := &sync.WaitGroup{}
	wg.Add(1)
	var newLeader RaftNode
	go func(w *sync.WaitGroup) {
		newLeader = waitForLeaderRole(nodes)
		w.Done()
	}(wg)

	leader.Stop()
	wg.Wait()

	oldLeader := leader

	t.Log(fmt.Sprintf("New leader - %s\n",newLeader.Id()))

	newNodes := make([]RaftNode,0)

	for _,node := range nodes {
		if node.IsRunning() {
			newNodes = append(newNodes,node)
		}
	}

	checkLeaderAmongN(newNodes,t)

	wg.Add(1)
	go func(w *sync.WaitGroup,n RaftNode) {
		defer w.Done()
		select {
		case <- n.RoleChange():
			return
		}
	}(wg,oldLeader)

	oldLeader.Start()

	wg.Wait()

	checkLeaderAmongN(nodes,t)

	
}
