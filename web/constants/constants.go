package constants

import (
	"github.com/AlyxPink/meower/web/grpc"
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	Web        *fiber.App
	GrpcClient *grpc.Client
}
