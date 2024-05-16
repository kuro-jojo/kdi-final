package main

import (
	"github.com/kuro-jojo/kdi-web/server"
)

func main() {
	// Load environment variables
	server.LoadEnv()

	// Initialize Server
	server.Init()
}
