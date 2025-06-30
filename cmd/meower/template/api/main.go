package main

import (
	"github.com/AlyxPink/meower/api/server"
)

const (
	apiEndpoint = "localhost:50051"
)

func main() {
	server.Serve()
}
