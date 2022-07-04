package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// UserRole holds the schema definition for the UserRole entity.
type UserRole struct {
	ent.Schema
}

// Fields of the UserRole.
func (UserRole) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.Int("role_id"),
	}
}

// Edges of the UserRole.
func (UserRole) Edges() []ent.Edge {
	return nil
}
