package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// BuildTask 构建任务模型
type BuildTask struct {
	ent.Schema
}

// Fields 构建任务字段定义
func (BuildTask) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").
			Values("pending", "running", "success", "failed"),
		field.String("log_path").Optional(),
		field.Int("duration").Optional().Comment("构建持续时间（秒）"),
		field.Time("started_at").Optional(),
		field.Time("finished_at").Optional().Nillable(),
		field.Time("created_at").Immutable(),
	}
}

// Edges 构建任务关系定义
func (BuildTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("builds").
			Unique().
			Required(),
	}
}