package data

import (
	"context"
	"kratos-todo-service/internal/biz"
	"kratos-todo-service/internal/data/ent"
	"kratos-todo-service/internal/data/ent/todo"

	"github.com/go-kratos/aip-go/ents"
	"github.com/google/uuid"
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

// FindByID implements [biz.TodoRepo].
func (r *todoRepo) FindByID(ctx context.Context, id uuid.UUID) (*biz.Todo, error) {
	po, err := r.data.db.Todo.Query().
		Where(todo.IDEQ(id), todo.StatusEQ(biz.TodoStatusActive)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrTodoNotFound
		}
		return nil, err
	}
	return toBiz(po), nil
}

func (r *todoRepo) ListTodos(ctx context.Context, opts ...biz.ListOption) ([]*biz.Todo, error) {
	options := biz.ListOptions{
		Limit: 20,
	}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrTodoInvalidArgument
	}

	pos, err := r.data.db.Todo.Query().
		Where(todo.StatusEQ(biz.TodoStatusActive)).
		Where(ents.ApplyFilter(options.Filter)).
		Order(ents.ApplyOrderBy(options.OrderBy), todo.ByID()).
		Offset(options.Offset).
		Limit(options.Limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	todos := make([]*biz.Todo, 0, len(pos))
	for _, po := range pos {
		todos = append(todos, toBiz(po))
	}
	return todos, nil
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

func (r *todoRepo) UpdateTodo(ctx context.Context, t *biz.Todo) (*biz.Todo, error) {
	po, err := r.data.db.Todo.UpdateOneID(t.ID).
		Where(todo.StatusEQ(biz.TodoStatusActive)).
		SetTitle(t.Title).
		SetContent(t.Content).
		SetCompleted(t.Completed).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrTodoNotFound
		}
		return nil, err
	}
	return toBiz(po), nil
}

func (r *todoRepo) DeleteTodo(ctx context.Context, id uuid.UUID) error {
	affected, err := r.data.db.Todo.Update().
		Where(todo.IDEQ(id), todo.StatusEQ(biz.TodoStatusActive)).
		SetStatus(biz.TodoStatusDeleted).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return biz.ErrTodoNotFound
	}
	return nil
}
