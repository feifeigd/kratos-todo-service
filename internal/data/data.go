package data

import (
	"github.com/google/wire"
	"kratos-todo-service/internal/data/ent"
)

var ProviderSet = wire.NewSet(NewData, NewTodoRepo)

type Data struct {
	db *ent.Client
}

// NewData opens the database client and returns it with a cleanup function.
func NewData() *Data {
	return &Data{}
}
