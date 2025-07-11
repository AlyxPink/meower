package main

import (
	"TEMPLATE_MODULE_PATH/api/server"
)

const (
	apiEndpoint = "localhost:50051"
)

func main() {
	server.Serve()
}
