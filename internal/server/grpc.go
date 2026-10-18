package server

import (
	v1 "kratos-todo-service/api/todo/v1"
	"kratos-todo-service/internal/conf"
	"kratos-todo-service/internal/service"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/transport/grpc"
)

func NewGRPCServer(c *conf.Server, todo *service.TodoService) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}

	conf := c.Grpc
	if conf.Network != "" {
		opts = append(opts, grpc.Network(conf.Network))
	}
	if conf.Addr != "" {
		opts = append(opts, grpc.Address(conf.Addr))
	}
	if conf.Timeout != nil {
		opts = append(opts, grpc.Timeout(conf.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterTodoServiceServer(srv, todo)
	return srv
}
