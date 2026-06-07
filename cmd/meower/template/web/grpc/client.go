package grpc

import (
	"os"

	meowV1 "TEMPLATE_MODULE_PATH/api/proto/meow/v1"
	userV1 "TEMPLATE_MODULE_PATH/api/proto/user/v1"
	"github.com/charmbracelet/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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
//
// The otelgrpc stats handler instruments every outgoing call: it creates a
// client span and injects the active trace context into the request metadata,
// so the API server (which runs the matching otelgrpc server handler) continues
// the same trace instead of starting a new one. This is what links a web trace
// to its downstream API spans — for it to work, callers must pass the
// request-scoped context that carries the web's span (c.Context() under
// otelfiber), not a fresh context.Background().
func NewClient() *Client {
	conn, err := grpc.NewClient(
		getApiEndpoint(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
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
