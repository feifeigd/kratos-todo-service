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
		IDMixin{},   // id
		TimeMixin{}, // created_at, updated_at
	}
}

// Fields of the Todo.
func (Todo) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").Default(""),
		field.String("content").Default(""),
		field.Bool("completed").Default(false),
	}
}
