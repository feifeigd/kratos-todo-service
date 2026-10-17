package server

import (
	v1 "kratos-todo-service/api/todo/v1"
	"kratos-todo-service/internal/conf"
	"kratos-todo-service/internal/service"

	"github.com/go-kratos/kratos/v3/transport/grpc"
)

func NewGRPCServer(c *conf.Server, todo *service.TodoService) *grpc.Server {
	srv := grpc.NewServer()
	v1.RegisterTodoServiceServer(srv, todo)
	return srv
}
