package service

import (
	"context"
	v1 "kratos-todo-service/api/todo/v1"
	"kratos-todo-service/internal/biz"
)

type TodoService struct {
	v1.UnimplementedTodoServiceServer
	uc *biz.TodoUsecase
}

// 实现 grpc 接口
var _ v1.TodoServiceServer = (*TodoService)(nil)

func NewTodoService(uc *biz.TodoUsecase) *TodoService {
	return &TodoService{uc: uc}
}

func(s* TodoService)CreateTodo(ctx context.Context, req *v1.CreateTodoRequest)(*v1.Todo, error){
	return nil, nil
}
