package main

import (
	"os"
	"fmt"
)




func main() {
	args := os.Args

	if len(args) != 2 {
		fmt.Println("Usage: server <port-number>")
		return
	}

	port := args[1]

	fmt.Println("Setting up...")
	
	h := NewHttpServer()

	fmt.Printf("Starting to listen on : %s", port)

	h.HandleRequests(port)
}
