package service

import (
	"context"
	v1 "kratos-todo-service/api/todo/v1"
	"kratos-todo-service/internal/biz"

	"github.com/google/uuid"
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

func (s *TodoService) CreateTodo(ctx context.Context, req *v1.CreateTodoRequest) (*v1.Todo, error) {
	todo, err := s.uc.CreateTodo(ctx, convertTodo(req.GetTodo()))
	if err != nil {
		return nil, err
	}
	return convertTodoReply(todo), nil
}

// convertTodo parses an incoming proto into a DO.
// Timestamps are omitted: they are server-assigned and read-only, so only the mutable fields cross over.
//
// The id is left at uuid.Nil when absent or unparseable.
// Callers that require a real id validate it with parseTodoID first; on the create path no id is supplied and the storage layer assigns one.
func convertTodo(in *v1.Todo) *biz.Todo {
	if in == nil {
		return nil
	}
	id, _ := uuid.Parse(in.GetId())
	return &biz.Todo{
		ID:        id,
		Title:     in.GetTitle(),
		Content:   in.GetContent(),
		Completed: in.GetCompleted(),
	}
}

func convertTodoReply(in *biz.Todo) *v1.Todo {
	if in == nil {
		return nil
	}
	return &v1.Todo{
		Id:        in.ID.String(),
		Title:     in.Title,
		Content:   in.Content,
		Completed: in.Completed,
		Status:    convertTodoStatus(in.Status),
	}
}

func convertTodoStatus(in biz.TodoStatus) v1.TodoStatus {
	switch in {
	case biz.TodoStatusActive:
		return v1.TodoStatus_TODO_STATUS_ACTIVE
	case biz.TodoStatusDeleted:
		return v1.TodoStatus_TODO_STATUS_DELETED
	default:
		return v1.TodoStatus_TODO_STATUS_UNSPECIFIED
	}
}
