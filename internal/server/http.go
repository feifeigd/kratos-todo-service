package server

import (
	v1 "kratos-todo-service/api/todo/v1" // api
	"kratos-todo-service/internal/conf"
	"kratos-todo-service/internal/service"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/middleware/validate"
	"github.com/go-kratos/kratos/v3/transport/http"

	"go.einride.tech/aip/fieldbehavior"
	"google.golang.org/protobuf/proto"
)

// 不能返回 nil
func NewHTTPServer(c *conf.Server, todo *service.TodoService) *http.Server {
	opts := []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			validate.Validator(func(req any) error {
				if msg, ok := req.(proto.Message); ok {
					if err := fieldbehavior.ValidateRequiredFields(msg); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	}

	httpConf := c.Http
	if httpConf.Network != "" {
		opts = append(opts, http.Network(httpConf.Network))
	}
	if httpConf.Addr != "" {
		opts = append(opts, http.Address(httpConf.Addr))
	}
	if httpConf.Timeout != nil {
		opts = append(opts, http.Timeout(httpConf.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)
	// 接口绑定实现
	v1.RegisterTodoServiceHTTPServer(srv, todo)
	return srv
}
