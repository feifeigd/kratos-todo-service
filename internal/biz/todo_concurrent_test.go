package biz_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"kratos-todo-service/internal/biz"

	"github.com/google/uuid"
)

// mockRepo 是一个内存版 TodoRepo。
// 它自身用 RWMutex 保护、读写都做拷贝，保证 mock 不产生数据竞争——
// 这样 race detector 报出的任何竞争都来自被测代码（TodoUsecase）本身。
type mockRepo struct {
	mu    sync.RWMutex
	store map[uuid.UUID]biz.Todo
}

func newMockRepo() *mockRepo {
	return &mockRepo{store: make(map[uuid.UUID]biz.Todo)}
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*biz.Todo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.store[id]
	if !ok || t.Status != biz.TodoStatusActive {
		return nil, biz.ErrTodoNotFound
	}
	cp := t
	return &cp, nil
}

func (m *mockRepo) ListTodos(_ context.Context, _ ...biz.ListOption) ([]*biz.Todo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*biz.Todo, 0, len(m.store))
	for _, t := range m.store {
		if t.Status != biz.TodoStatusActive {
			continue
		}
		cp := t
		out = append(out, &cp)
	}
	return out, nil
}

func (m *mockRepo) CreateTodo(_ context.Context, t *biz.Todo) (*biz.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *t
	cp.ID = uuid.New()
	cp.Status = biz.TodoStatusActive
	m.store[cp.ID] = cp
	return &cp, nil
}

func (m *mockRepo) UpdateTodo(_ context.Context, t *biz.Todo) (*biz.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.store[t.ID]
	if !ok || cur.Status != biz.TodoStatusActive {
		return nil, biz.ErrTodoNotFound
	}
	cp := *t
	cp.Status = biz.TodoStatusActive
	m.store[cp.ID] = cp
	return &cp, nil
}

func (m *mockRepo) DeleteTodo(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.store[id]
	if !ok || cur.Status != biz.TodoStatusActive {
		return biz.ErrTodoNotFound
	}
	cur.Status = biz.TodoStatusDeleted
	m.store[id] = cur
	return nil
}

// TestTodoUsecaseConcurrent 模拟 kratos 运行时：单个 TodoUsecase 单例被大量
// goroutine 并发调用。用 go test -race 运行时，任何共享状态竞争都会报错。
func TestTodoUsecaseConcurrent(t *testing.T) {
	uc := biz.NewTodoUsecase(newMockRepo())
	ctx := context.Background()

	const workers = 64
	const iterations = 200

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(w int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				title := fmt.Sprintf("todo-%d-%d", w, i)

				created, err := uc.CreateTodo(ctx, &biz.Todo{Title: title})
				if err != nil {
					t.Errorf("CreateTodo: %v", err)
					return
				}

				if _, err := uc.GetTodo(ctx, created.ID); err != nil {
					t.Errorf("GetTodo: %v", err)
					return
				}

				created.Completed = i%2 == 0
				if _, err := uc.UpdateTodo(ctx, created); err != nil {
					t.Errorf("UpdateTodo: %v", err)
					return
				}

				if _, err := uc.ListTodos(ctx, biz.ListLimit(10)); err != nil {
					t.Errorf("ListTodos: %v", err)
					return
				}

				if i%3 == 0 {
					if err := uc.DeleteTodo(ctx, created.ID); err != nil {
						t.Errorf("DeleteTodo: %v", err)
						return
					}
				}
			}
		}(w)
	}
	wg.Wait()
}
