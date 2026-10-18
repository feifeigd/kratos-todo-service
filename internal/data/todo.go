package data

import (
	"context"
	"kratos-todo-service/internal/biz"
	"kratos-todo-service/internal/data/ent"
)

// po 在 ent/schema 中定义
func toBiz(po *ent.Todo) *biz.Todo {
	if po == nil {
		return nil
	}
	return &biz.Todo{
		ID:        po.ID,
		Title:     po.Title,
		Content:   po.Content,
		Completed: po.Completed,
		Status:    po.Status,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

type todoRepo struct {
	data *Data
}

func NewTodoRepo(data *Data) biz.TodoRepo {
	return &todoRepo{data: data}
}

func (r *todoRepo) CreateTodo(ctx context.Context, t *biz.Todo) (*biz.Todo, error) {
	po, err := r.data.db.Todo.Create().
		SetTitle(t.Title).
		SetContent(t.Content).
		SetCompleted(t.Completed).
		SetStatus(biz.TodoStatusActive).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return toBiz(po), nil
}
