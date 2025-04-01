package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User 用户模型
type User struct {
	ent.Schema
}

// Fields 用户字段定义
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Unique(),
		field.String("password").Sensitive(),
		field.String("gitee_token").Optional().Sensitive(),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

// Edges 用户关系定义
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("projects", Project.Type),
	}
}