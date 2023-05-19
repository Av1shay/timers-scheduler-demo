package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

type Task struct {
	ent.Schema
}

func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.Time("dueDate"),
		field.String("webhookUrl"),
		field.Enum("status").Values("pending", "running", "done").Default("pending"),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Task) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("dueDate"),
		index.Fields("status"),
		index.Fields("dueDate", "status"),
	}
}

func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("histories", TaskHistory.Type),
	}
}
