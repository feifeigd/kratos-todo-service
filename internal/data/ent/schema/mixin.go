package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// IDMixin adds the standard application-generated UUID primary key to schemas.
type IDMixin struct {
	mixin.Schema
}

// Fields of the IDMixin.
func (IDMixin) Fields() []ent.Field {
	return []ent.Field{
		// No SchemaType: ent already maps TypeUUID per dialect (char(36) on MySQL, native uuid on Postgres),
		// so the GO side stays uuid.UUID everywhere without pinning a column type.
		field.UUID("id", uuid.Nil).Default(newUUID).Immutable(),
	}
}

func newUUID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
