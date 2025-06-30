package main

import (
	"{{.ModulePath}}/api/server"
)

const (
	apiEndpoint = "localhost:50051"
)

func main() {
	server.Serve()
}
