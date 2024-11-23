package grpc

import (
	"os"

	pb "github.com/AlyxPink/meower/api/proto"
	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	apiEndpoint = "localhost:50051"
)

type Client struct {
	MeowService pb.MeowServiceClient
	conn        *grpc.ClientConn
}

// NewClient initializes and returns a new gRPC client for our services API.
func NewClient() *Client {
	conn, err := grpc.NewClient(getApiEndpoint(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	client := &Client{
		MeowService: pb.NewMeowServiceClient(conn),
		conn:        conn,
	}

	return client
}

func getApiEndpoint() string {
	if os.Getenv("API_ENDPOINT") != "" {
		return os.Getenv("API_ENDPOINT")
	}
	return apiEndpoint
}
