package grpc

import (
	"os"

	meowV1 "TEMPLATE_MODULE_PATH/api/proto/meow/v1"
	userV1 "TEMPLATE_MODULE_PATH/api/proto/user/v1"
	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	apiEndpoint = "localhost:50051"
)

type Client struct {
	MeowService meowV1.MeowServiceClient
	UserService userV1.UserServiceClient
	conn        *grpc.ClientConn
}

// NewClient initializes and returns a new gRPC client for our services API.
func NewClient() *Client {
	conn, err := grpc.NewClient(getApiEndpoint(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	client := &Client{
		MeowService: meowV1.NewMeowServiceClient(conn),
		UserService: userV1.NewUserServiceClient(conn),
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
