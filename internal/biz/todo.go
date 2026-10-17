package biz

import (
	"context"
	v1 "kratos-todo-service/api/todo/v1"
	"strings"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/google/uuid"
)

var (
	ErrTodoNotFound        = errors.NotFound(v1.ErrorReason_TODO_NOT_FOUND.String(), "todo not found")
	ErrTodoInvalidArgument = errors.BadRequest(v1.ErrorReason_TODO_INVALID_ARGUMENT.String(), "invalid todo argument")
)

// 枚举
// TodoStatus is the lifecycle state of a todo.
type TodoStatus int32

const (
	TodoStatusUnspecified TodoStatus = 0
	TodoStatusActive      TodoStatus = 1
	TodoStatusDeleted     TodoStatus = 2
)

type (
	// Todo is a Todo model.
	Todo struct {
		ID        uuid.UUID
		Title     string
		Content   string
		Completed bool
		Status    TodoStatus
	}

	TodoRepo interface {
		CreateTodo(context.Context, *Todo) (*Todo, error)
	}

	TodoUsecase struct {
		repo TodoRepo
	}
)

func NewTodoUsecase(repo TodoRepo) *TodoUsecase {
	return &TodoUsecase{repo: repo}
}

func (uc *TodoUsecase) CreateTodo(ctx context.Context, todo *Todo) (*Todo, error) {
	if err := validateTodo(todo); err != nil {
		return nil, err
	}
	return uc.repo.CreateTodo(ctx, todo)
}

func validateTodo(todo *Todo) error {
	if todo == nil || strings.TrimSpace(todo.Title) == "" {
		return ErrTodoInvalidArgument
	}
	return nil
}
