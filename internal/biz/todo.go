package biz

import (
	"context"
	v1 "kratos-todo-service/api/todo/v1"
	"strings"
	"time"

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
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	TodoRepo interface {
		FindByID(context.Context, uuid.UUID) (*Todo, error)
		ListTodos(context.Context, ...ListOption) ([]*Todo, error)
		CreateTodo(context.Context, *Todo) (*Todo, error)
		UpdateTodo(context.Context, *Todo) (*Todo, error)
		DeleteTodo(context.Context, uuid.UUID) error
	}

	TodoUsecase struct {
		repo TodoRepo
	}
)

func NewTodoUsecase(repo TodoRepo) *TodoUsecase {
	return &TodoUsecase{repo: repo}
}

func (uc *TodoUsecase) GetTodo(ctx context.Context, id uuid.UUID) (*Todo, error) {
	if id == uuid.Nil {
		return nil, ErrTodoInvalidArgument
	}
	return uc.repo.FindByID(ctx, id)
}

func (uc *TodoUsecase) ListTodos(ctx context.Context, opts ...ListOption) ([]*Todo, error) {
	return uc.repo.ListTodos(ctx, opts...)
}

func (uc *TodoUsecase) CreateTodo(ctx context.Context, todo *Todo) (*Todo, error) {
	if err := validateTodo(todo); err != nil {
		return nil, err
	}
	return uc.repo.CreateTodo(ctx, todo)
}

func (uc *TodoUsecase) UpdateTodo(ctx context.Context, todo *Todo) (*Todo, error) {
	if todo == nil || todo.ID == uuid.Nil {
		return nil, ErrTodoInvalidArgument
	}
	if err := validateTodo(todo); err != nil {
		return nil, err
	}
	return uc.repo.UpdateTodo(ctx, todo)
}

func (uc *TodoUsecase) DeleteTodo(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrTodoInvalidArgument
	}

	return uc.repo.DeleteTodo(ctx, id)
}

func validateTodo(todo *Todo) error {
	if todo == nil || strings.TrimSpace(todo.Title) == "" {
		return ErrTodoInvalidArgument
	}
	return nil
}
