package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

type TaskHistory struct {
	ent.Schema
}

func (TaskHistory) Fields() []ent.Field {
	return []ent.Field{
		field.String("error").Optional().Nillable(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (TaskHistory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", Task.Type).
			Ref("histories").
			Unique(),
	}
}
