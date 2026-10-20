package service

import (
	"context"
	v1 "kratos-todo-service/api/todo/v1"
	"kratos-todo-service/internal/biz"
	"testing"
	"time"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func newTestTodoService() *TodoService {
	uc := biz.NewTodoUsecase(newFakeTodoRepo())
	return NewTodoService(uc)
}

type fakeTodoRepo struct {
	order []uuid.UUID
	todos map[uuid.UUID]*biz.Todo
}

// CreateTodo implements [biz.TodoRepo].
func (r *fakeTodoRepo) CreateTodo(_ context.Context, todo *biz.Todo) (*biz.Todo, error) {
	now := time.Now()
	todo = cloneTodo(todo)
	todo.ID = uuid.Must(uuid.NewV7())
	todo.Status = biz.TodoStatusActive
	todo.CreatedAt = now
	todo.UpdatedAt = now
	r.order = append(r.order, todo.ID)
	r.todos[todo.ID] = cloneTodo(todo)
	return cloneTodo(todo), nil
}

// DeleteTodo implements [biz.TodoRepo].
func (r *fakeTodoRepo) DeleteTodo(_ context.Context, id uuid.UUID) error {
	if _, ok := r.todos[id]; !ok {
		return biz.ErrTodoNotFound
	}
	delete(r.todos, id)
	return nil
}

// FindByID implements [biz.TodoRepo].
func (r *fakeTodoRepo) FindByID(_ context.Context, id uuid.UUID) (*biz.Todo, error) {
	todo, ok := r.todos[id]
	if !ok {
		return nil, biz.ErrTodoNotFound
	}
	return cloneTodo(todo), nil
}

func cloneTodo(todo *biz.Todo) *biz.Todo {
	if todo == nil {
		return nil
	}
	clone := *todo
	return &clone
}

// ListTodos implements [biz.TodoRepo].
func (r *fakeTodoRepo) ListTodos(_ context.Context, opts ...biz.ListOption) ([]*biz.Todo, error) {
	options := biz.ListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrTodoInvalidArgument
	}
	todos := make([]*biz.Todo, 0, len(r.todos))
	for _, id := range r.order {
		if todo, ok := r.todos[id]; ok {
			todos = append(todos, cloneTodo(todo))
		}
	}
	if options.Offset >= len(todos) {
		return []*biz.Todo{}, nil
	}
	end := min(options.Offset+options.Limit, len(todos))
	return todos[options.Offset:end], nil
}

// UpdateTodo implements [biz.TodoRepo].
func (r *fakeTodoRepo) UpdateTodo(_ context.Context, todo *biz.Todo) (*biz.Todo, error) {
	current, ok := r.todos[todo.ID]
	if !ok {
		return nil, biz.ErrTodoNotFound
	}
	updated := cloneTodo(todo)
	updated.Status = current.Status
	updated.CreatedAt = current.CreatedAt
	updated.UpdatedAt = time.Now()
	r.todos[updated.ID] = cloneTodo(updated)
	return cloneTodo(updated), nil
}

func newFakeTodoRepo() biz.TodoRepo {
	return &fakeTodoRepo{todos: make(map[uuid.UUID]*biz.Todo)}
}

func TestTodoServiceCRUD(t *testing.T) {
	ctx := context.Background()
	svc := newTestTodoService()

	created, err := svc.CreateTodo(ctx, &v1.CreateTodoRequest{
		Todo: &v1.Todo{
			Title:     "write tests",
			Content:   "cover todo CRUD",
			Completed: false,
		},
	})
	if err != nil {
		t.Fatalf("CreatedTodo() error = %v", err)
	}
	if _, err := uuid.Parse(created.GetId()); err != nil {
		t.Fatalf("CreatedTodo() id = %q, want a UUID", created.GetId())
	}
	if created.GetCreatedAt() == nil || created.GetUpdatedAt() == nil {
		t.Fatal("CreatedTodos() did not set timestamps")
	}
	if created.GetStatus() != v1.TodoStatus_TODO_STATUS_ACTIVE {
		t.Fatalf("CreatedTodo() status = %v, want active", created.GetStatus())
	}
	got, err := svc.GetTodo(ctx, &v1.GetTodoRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetTodo() error = %v", err)
	}
	if got.GetTitle() != "write tests" || got.GetContent() != "cover todo CRUD" {
		t.Fatalf("GetTodo() = %+v, want created todo", got)
	}

	updated, err := svc.UpdateTodo(ctx, &v1.UpdateTodoRequest{
		Todo: &v1.Todo{
			Id:        created.GetId(),
			Title:     "write service tests",
			Completed: true,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"title", "completed"}},
	})
	if err != nil {
		t.Fatalf("UpdatedTodo() error = %v", err)
	}
	if updated.GetTitle() != "write service tests" || !updated.GetCompleted() {
		t.Fatalf("UpdateTodo() = %+v, want updated title and completed", updated)
	}
	if updated.GetContent() != "cover todo CRUD" {
		t.Fatalf("UpdateTodo content = %q, want original content", updated.GetContent())
	}
	if updated.GetStatus() != v1.TodoStatus_TODO_STATUS_ACTIVE {
		t.Fatalf("UpdateTodo() status = %v, want active", updated.GetStatus())
	}

	if _, err := svc.DeleteTodo(ctx, &v1.DeleteTodoRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeleteTodo() error = %v", err)
	}
	if _, err := svc.GetTodo(ctx, &v1.GetTodoRequest{Id: created.GetId()}); !kratoserrors.IsNotFound(err) {
		t.Fatalf("GetTodo() after deleted error = %v, want not found", err)
	}
}
