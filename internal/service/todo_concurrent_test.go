package service_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"

	v1 "kratos-todo-service/api/todo/v1"
	"kratos-todo-service/internal/biz"
	"kratos-todo-service/internal/service"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// mockRepo 是内存版 TodoRepo，自身用 RWMutex 保护、读写拷贝，
// 保证 race detector 报出的竞争只可能来自被测的 service/biz 代码。
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

func (m *mockRepo) ListTodos(_ context.Context, opts ...biz.ListOption) ([]*biz.Todo, error) {
	options := biz.ListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*biz.Todo, 0, len(m.store))
	for _, t := range m.store {
		if t.Status != biz.TodoStatusActive {
			continue
		}
		if options.Offset > 0 {
			options.Offset--
			continue
		}
		if len(out) >= options.Limit {
			break
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

// newTestClient 用 bufconn 在内存里起一个真实 gRPC server，
// 挂的是完整的 service -> biz -> repo 链路，和 kratos 运行时同一套代码路径。
func newTestClient(t *testing.T) (v1.TodoServiceClient, func()) {
	t.Helper()

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	v1.RegisterTodoServiceServer(srv, service.NewTodoService(biz.NewTodoUsecase(newMockRepo())))
	go func() { _ = srv.Serve(lis) }()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial bufconn: %v", err)
	}

	cleanup := func() {
		conn.Close()
		srv.Stop()
		lis.Close()
	}
	return v1.NewTodoServiceClient(conn), cleanup
}

// TestTodoServiceConcurrentE2E 端到端并发测试：多个 goroutine 经真实 gRPC
// 调用并发打 unary（CRUD/List）、server 流（WatchTodos）和双向流（SyncTodos）。
// 用 go test -race 运行时，任何共享状态竞争都会报错。
func TestTodoServiceConcurrentE2E(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	ctx := context.Background()

	const workers = 16
	const iterations = 20

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(w int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				title := fmt.Sprintf("todo-%d-%d", w, i)

				// unary: create
				created, err := client.CreateTodo(ctx, &v1.CreateTodoRequest{
					Todo: &v1.Todo{Title: title, Content: "content"},
				})
				if err != nil {
					t.Errorf("CreateTodo: %v", err)
					return
				}

				// unary: get
				if _, err := client.GetTodo(ctx, &v1.GetTodoRequest{Id: created.GetId()}); err != nil {
					t.Errorf("GetTodo: %v", err)
					return
				}

				// unary: update（field mask 路径）
				if _, err := client.UpdateTodo(ctx, &v1.UpdateTodoRequest{
					Todo:       &v1.Todo{Id: created.GetId(), Completed: true},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"completed"}},
				}); err != nil {
					t.Errorf("UpdateTodo: %v", err)
					return
				}

				// unary: list（filtering/ordering/pagination 解析路径）
				if _, err := client.ListTodos(ctx, &v1.ListTodosRequest{PageSize: 10}); err != nil {
					t.Errorf("ListTodos: %v", err)
					return
				}

				// server stream: watch
				watch, err := client.WatchTodos(ctx, &v1.WatchTodosRequest{PageSize: 10})
				if err != nil {
					t.Errorf("WatchTodos: %v", err)
					return
				}
				for {
					if _, err := watch.Recv(); err == io.EOF {
						break
					} else if err != nil {
						t.Errorf("WatchTodos Recv: %v", err)
						return
					}
				}

				// bidi stream: sync（create + delete 各一条）
				syncStream, err := client.SyncTodos(ctx)
				if err != nil {
					t.Errorf("SyncTodos: %v", err)
					return
				}
				if err := syncStream.Send(&v1.SyncTodoRequest{
					Action: "create",
					Todo:   &v1.Todo{Title: title + "-sync"},
				}); err != nil {
					t.Errorf("SyncTodos Send create: %v", err)
					return
				}
				event, err := syncStream.Recv()
				if err != nil {
					t.Errorf("SyncTodos Recv: %v", err)
					return
				}
				if err := syncStream.Send(&v1.SyncTodoRequest{
					Action: "delete",
					Id:     event.GetTodo().GetId(),
				}); err != nil {
					t.Errorf("SyncTodos Send delete: %v", err)
					return
				}
				if _, err := syncStream.Recv(); err != nil {
					t.Errorf("SyncTodos Recv delete: %v", err)
					return
				}
				if err := syncStream.CloseSend(); err != nil {
					t.Errorf("SyncTodos CloseSend: %v", err)
					return
				}

				// unary: delete
				if i%3 == 0 {
					if _, err := client.DeleteTodo(ctx, &v1.DeleteTodoRequest{Id: created.GetId()}); err != nil {
						t.Errorf("DeleteTodo: %v", err)
						return
					}
				}
			}
		}(w)
	}
	wg.Wait()
}
