package schema

import (
	"regexp"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		// UUID stored as string to match models.User.ID
		field.String("id").
			NotEmpty().
			Immutable(), // Primary key, never changes after creation

		field.String("username").
			MaxLen(30).
			NotEmpty().
			Match(regexp.MustCompile(`^[a-z0-9_]{3,30}$`)), // Match DB CHECK constraint

		// Sensitive: Ent won't include this field in logs or String() output
		field.String("password_hash").
			NotEmpty().
			Sensitive(),

		field.Time("created_at").
			Default(time.Now).
			Immutable(), // Never changes after creation

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now), // Auto-updated on every save
	}
}

// Edges of the User.
// These define relationships — Ent uses these to generate JOIN helpers.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		// Messages this user sent (one user → many direct messages as sender)
		edge.To("sent_messages", DirectMessage.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),

		// Messages this user received (one user → many direct messages as recipient)
		edge.To("received_messages", DirectMessage.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),

		// Groups this user created (one user → many groups)
		edge.To("created_groups", Group.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),

		// Group memberships (one user → many group_members rows)
		edge.To("memberships", GroupMember.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),

		// Group messages this user sent (one user → many group_messages)
		edge.To("sent_group_messages", GroupMessage.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username").Unique(), // Enforces username uniqueness
	}
}

// Annotations tell Ent which table name to use.
// Without this, Ent would look for "users" anyway (plural), but being
// explicit means we're guaranteed to match the existing table.
func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "users"},
	}
}
