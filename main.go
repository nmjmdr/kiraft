package main

import (
	"os"
	"bufio"
)

func main() {
	s := NewSimulator(1)

	s.Start()

	bufio.NewScanner(os.Stdin).Scan()
	
	s.Stop()
	
}
