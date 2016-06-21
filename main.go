package main

import (
	"os"
	"bufio"
)

func main() {
	s := NewSimulator(5)

	s.Start()

	bufio.NewScanner(os.Stdin).Scan()
	
	s.Stop()
	
}
