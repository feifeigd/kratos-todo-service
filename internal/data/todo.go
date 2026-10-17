package data

import (
	"context"
	"kratos-todo-service/internal/biz"
)

type todoRepo struct {
}

func NewTodoRepo() biz.TodoRepo {
	return &todoRepo{}
}

func (r *todoRepo) CreateTodo(ctx context.Context, t *biz.Todo) (*biz.Todo, error) {
	return nil, nil
}
