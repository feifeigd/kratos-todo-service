package data

import (
	"context"
	"kratos-todo-service/internal/conf"
	"kratos-todo-service/internal/data/ent"

	"github.com/go-kratos/kratos/v3/log"
	_ "github.com/go-sql-driver/mysql" // 导入数据库驱动
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewData, NewTodoRepo)

type Data struct {
	db *ent.Client
}

// NewData opens the database client and returns it with a cleanup function.
// 创建数据库连接
// 构建函数的返回值:
// 1: *
// 2: *, error
// 3: *, cleanup func(), error
func NewData(c *conf.Data) (*Data, func(), error) {
	dc := c.GetDatabase()
	db, err := ent.Open(dc.GetDriver(), dc.GetSource())
	if err != nil {
		return nil, nil, err
	}

	if dc.GetDebug() {
		db = db.Debug()
	}

	// Auto migration is a convenience for local development.
	// In production, apply schema changes as a separate reviewed step instead.
	if dc.GetAutoMigrate() {
		if err := db.Schema.Create(context.Background()); err != nil {
			db.Close()
			return nil, nil, err
		}
	}
	// 析构函数
	cleanup := func() {
		log.Info("Closing the data resources.")
		if err := db.Close(); err != nil {
			log.Error("failed closing the database", "err", err)
		}
	}
	return &Data{db: db}, cleanup, nil
}
