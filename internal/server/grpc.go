package server

import (
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"kratos-todo-service/internal/conf"
)

func NewGRPCServer(c *conf.Server) *grpc.Server {
	srv := grpc.NewServer()

	return srv
}
