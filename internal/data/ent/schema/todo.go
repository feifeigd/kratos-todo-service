package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// 数据库表 todo 定义
type Todo struct {
	ent.Schema
}

func (Todo) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{}, // id
	}
}

// Fields of the Todo.
func (Todo) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").Default(""),
	}
}
