package schema

import (
	"kratos-todo-service/internal/biz"

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
		// status marks the row lifecycle: deletes flip it to deleted instead of removing the row,
		// so every read filters on active.
		// Stored as an integer bound to the domain type,
		// whose values match the api enum.
		field.Int32("status").GoType(biz.TodoStatus(biz.TodoStatusUnspecified)).Default(int32(biz.TodoStatusActive)),
	}
}
