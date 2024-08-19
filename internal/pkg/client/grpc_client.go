package client

import (
	"os"

	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	apiEndpoint = "localhost:50051"
)

// NewClient initializes and returns a new gRPC client for our services API.
func NewClient() *grpc.ClientConn {
	conn, err := grpc.Dial(getApiEndpoint(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	return conn
}

func getApiEndpoint() string {
	if os.Getenv("API_ENDPOINT") != "" {
		return os.Getenv("API_ENDPOINT")
	}
	return apiEndpoint
}
