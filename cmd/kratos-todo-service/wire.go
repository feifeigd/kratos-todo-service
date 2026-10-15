package main

import(
	"log/slog"
	"kratos-todo-service/biz"
	"kratos-todo-service/conf"
	"kratos-todo-service/data"
	"kratos-todo-service/server"
	"kratos-todo-service/service"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Data, *slog.Logger)(*kratos.App, func(), err){
	panic(wire.Build(
		server.ProviderSet, 
		data.ProviderSet, 
		biz.ProviderSet, 
		service.ProviderSet, 
		newApp))
}
