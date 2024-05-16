package main

import (
	"github.com/kuro-jojo/kdi-k8s/server"
)

func main() {
	// Load environment variables
	server.LoadEnv()
	// Initialize Server
	server.Init()
}
