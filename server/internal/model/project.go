// Code generated by ent, DO NOT EDIT.

package model

import (
	"fmt"
	"minidevops/internal/model/project"
	"minidevops/internal/model/user"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

// Project is the model entity for the Project schema.
type Project struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// RepoURL holds the value of the "repo_url" field.
	RepoURL string `json:"repo_url,omitempty"`
	// Branch holds the value of the "branch" field.
	Branch string `json:"branch,omitempty"`
	// BuildCmd holds the value of the "build_cmd" field.
	BuildCmd string `json:"build_cmd,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// LastBuildAt holds the value of the "last_build_at" field.
	LastBuildAt *time.Time `json:"last_build_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the ProjectQuery when eager-loading is set.
	Edges         ProjectEdges `json:"edges"`
	user_projects *int
	selectValues  sql.SelectValues
}

// ProjectEdges holds the relations/edges for other nodes in the graph.
type ProjectEdges struct {
	// Owner holds the value of the owner edge.
	Owner *User `json:"owner,omitempty"`
	// Builds holds the value of the builds edge.
	Builds []*BuildTask `json:"builds,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// OwnerOrErr returns the Owner value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e ProjectEdges) OwnerOrErr() (*User, error) {
	if e.loadedTypes[0] {
		if e.Owner == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: user.Label}
		}
		return e.Owner, nil
	}
	return nil, &NotLoadedError{edge: "owner"}
}

// BuildsOrErr returns the Builds value or an error if the edge
// was not loaded in eager-loading.
func (e ProjectEdges) BuildsOrErr() ([]*BuildTask, error) {
	if e.loadedTypes[1] {
		return e.Builds, nil
	}
	return nil, &NotLoadedError{edge: "builds"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Project) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case project.FieldID:
			values[i] = new(sql.NullInt64)
		case project.FieldName, project.FieldRepoURL, project.FieldBranch, project.FieldBuildCmd:
			values[i] = new(sql.NullString)
		case project.FieldCreatedAt, project.FieldUpdatedAt, project.FieldLastBuildAt:
			values[i] = new(sql.NullTime)
		case project.ForeignKeys[0]: // user_projects
			values[i] = new(sql.NullInt64)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Project fields.
func (pr *Project) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case project.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			pr.ID = int(value.Int64)
		case project.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				pr.Name = value.String
			}
		case project.FieldRepoURL:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field repo_url", values[i])
			} else if value.Valid {
				pr.RepoURL = value.String
			}
		case project.FieldBranch:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field branch", values[i])
			} else if value.Valid {
				pr.Branch = value.String
			}
		case project.FieldBuildCmd:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field build_cmd", values[i])
			} else if value.Valid {
				pr.BuildCmd = value.String
			}
		case project.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				pr.CreatedAt = value.Time
			}
		case project.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				pr.UpdatedAt = value.Time
			}
		case project.FieldLastBuildAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field last_build_at", values[i])
			} else if value.Valid {
				pr.LastBuildAt = new(time.Time)
				*pr.LastBuildAt = value.Time
			}
		case project.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field user_projects", value)
			} else if value.Valid {
				pr.user_projects = new(int)
				*pr.user_projects = int(value.Int64)
			}
		default:
			pr.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Project.
// This includes values selected through modifiers, order, etc.
func (pr *Project) Value(name string) (ent.Value, error) {
	return pr.selectValues.Get(name)
}

// QueryOwner queries the "owner" edge of the Project entity.
func (pr *Project) QueryOwner() *UserQuery {
	return NewProjectClient(pr.config).QueryOwner(pr)
}

// QueryBuilds queries the "builds" edge of the Project entity.
func (pr *Project) QueryBuilds() *BuildTaskQuery {
	return NewProjectClient(pr.config).QueryBuilds(pr)
}

// Update returns a builder for updating this Project.
// Note that you need to call Project.Unwrap() before calling this method if this Project
// was returned from a transaction, and the transaction was committed or rolled back.
func (pr *Project) Update() *ProjectUpdateOne {
	return NewProjectClient(pr.config).UpdateOne(pr)
}

// Unwrap unwraps the Project entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (pr *Project) Unwrap() *Project {
	_tx, ok := pr.config.driver.(*txDriver)
	if !ok {
		panic("model: Project is not a transactional entity")
	}
	pr.config.driver = _tx.drv
	return pr
}

// String implements the fmt.Stringer.
func (pr *Project) String() string {
	var builder strings.Builder
	builder.WriteString("Project(")
	builder.WriteString(fmt.Sprintf("id=%v, ", pr.ID))
	builder.WriteString("name=")
	builder.WriteString(pr.Name)
	builder.WriteString(", ")
	builder.WriteString("repo_url=")
	builder.WriteString(pr.RepoURL)
	builder.WriteString(", ")
	builder.WriteString("branch=")
	builder.WriteString(pr.Branch)
	builder.WriteString(", ")
	builder.WriteString("build_cmd=")
	builder.WriteString(pr.BuildCmd)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(pr.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(pr.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	if v := pr.LastBuildAt; v != nil {
		builder.WriteString("last_build_at=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteByte(')')
	return builder.String()
}

// Projects is a parsable slice of Project.
type Projects []*Project
