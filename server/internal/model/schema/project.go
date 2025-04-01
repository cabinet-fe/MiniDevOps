package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Project 项目模型
type Project struct {
	ent.Schema
}

// Fields 项目字段定义
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("repo_url"),
		field.String("branch").Default("master"),
		field.String("build_cmd"),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
		field.Time("last_build_at").Optional().Nillable(),
	}
}

// Edges 项目关系定义
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("projects").
			Unique().
			Required(),
		edge.To("builds", BuildTask.Type),
	}
}