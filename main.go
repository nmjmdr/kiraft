package main

import (
	"os"
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

var numNodes int

func main() {

	args := os.Args

	if len(args) != 2 {
		printHelp()
		return
	}


	num,err := strconv.Atoi(args[1])

	if err != nil {
		printHelp()
		return
	}

	numNodes = num
	s := NewSimulator(numNodes)

	s.Start()

	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		
		ok := scanner.Scan()
		if !ok {
			continue
		}
		command := scanner.Text()		
		if command == "quit" {
			s.Stop()
			return
		}
		handleCommand(s,command)
	}

	fmt.Println("Node States:")
	states := s.GetStates()

	for _,state := range states {
		fmt.Println(state)
	}
	
	s.Stop()
	
}

func handleCommand(s Simulator,command string) {

	splits := strings.Split(command," ")

	if  splits[0] == "get" {
		states:= s.GetStates()
		fmt.Println("Node States:")
		for index,state := range states {
			running,_ := s.IsRunning(index)
			if running {
				fmt.Println(state)
			} else {
				fmt.Printf("Node %d - is stopped\n",index)
			}
		}
		return
	}

	if len(splits) != 2 {
		fmt.Println("start/stop node-index")
		return
	}

	nodeIndex, err := strconv.Atoi(splits[1])

	if err != nil {
		fmt.Println("start/stop node-index")
		return
	}

	if nodeIndex < 0 || nodeIndex >= numNodes {
		fmt.Printf("Node-index should be between 0 to %d (inclusive)\n",numNodes-1)
		return
	}

		
	switch splits[0] {
	case "start":
		s.StartNode(nodeIndex)
	case "stop":
		s.StopNode(nodeIndex)		
	default:		
		fmt.Println("<start/stop node-index</<get>")
		return
	}

	
}


func printHelp() {
	fmt.Println("Usage: kiraft <number-of-nodes>")
	fmt.Println("start node-index - to start a stopped node")
	fmt.Println("stop node-index - to stop a node")
	fmt.Println("get to get status of all nodes")
	fmt.Println("quit")
}
