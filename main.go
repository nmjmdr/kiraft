package main

import (
	"os"
	"bufio"
	"fmt"
)

func main() {
	s := NewSimulator(5)

	s.Start()

	bufio.NewScanner(os.Stdin).Scan()

	fmt.Println("Node States:")
	states := s.GetStates()

	for _,state := range states {
		fmt.Println(state)
	}
	
	s.Stop()
	
}
