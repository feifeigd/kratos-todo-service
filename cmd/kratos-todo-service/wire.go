// 这行注释是一个构建约束，它告诉 go build 只有在满足条件 wireinject 的情况下才应该构建这个文件。
// wireinject 是一个特殊的标签，用于指示 wire 工具处理这个文件。
//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

// 执行命令生成 wire_gen.go
// wire ./...
// 或者
// go generate ./...

package main

import (
	"log/slog"
	// "kratos-todo-service/internal/biz"
	"kratos-todo-service/internal/conf"
	// "kratos-todo-service/internal/data"
	"kratos-todo-service/internal/server"
	"kratos-todo-service/internal/service"

	"github.com/go-kratos/kratos/v3" // kratos
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *slog.Logger) (app *kratos.App, cleanup func(), err error) {
	// wireApp 函数中的参数，会按需生成到 wire.Build 函数列表调用, 看 wire_gen.go
	panic(wire.Build(
		server.ProviderSet,
		// data.ProviderSet,
		// biz.ProviderSet,
		service.ProviderSet,
		newApp))
}
