package data

import (
	"kratos-todo-service/internal/conf"
	"kratos-todo-service/internal/data/ent"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewData, NewTodoRepo)

type Data struct {
	db *ent.Client
}

// NewData opens the database client and returns it with a cleanup function.
func NewData(c *conf.Data) *Data {
	dc := c.GetDatabase()
	db, err := ent.Open(dc.GEtDriver(), dc.GetSource())
	if err != nil {
		return nil
	}

	return &Data{db: db}
}
