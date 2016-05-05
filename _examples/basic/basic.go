package main

import (
	"fmt"

	"gopkg.in/vinxi/vinxi.v0"
)

func main() {
	vs := vinxi.NewServer(vinxi.ServerOptions{Port: 3100})

	// Forward all the traffic to httpbin.org
	vs.Forward("http://httpbin.org")

	fmt.Printf("Server listening on port: %d\n", 3100)
	err := vs.Listen()
	if err != nil {
		fmt.Errorf("Error: %s\n", err)
	}
}
