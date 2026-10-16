package service

import (
	
	"context"
	"kratos-todo-service/api/todo/v1"
)
type TodoService struct{

}

// 实现 grpc 接口
// var _ v1.TodoService = (*TodoService)(nil)

func NewTodoService()*TodoService{
	return &TodoService{}
}

func(s* TodoService)CreateTodo(ctx context.Context, req *v1.CreateTodoRequest)(*v1.Todo, error){
	return nil, nil
}
