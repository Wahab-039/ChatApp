package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Group holds the schema definition for the Group entity.
type Group struct {
	ent.Schema
}

// Fields of the Group.
func (Group) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Immutable(),

		field.String("name").
			MaxLen(100).
			NotEmpty(),

		// created_by stores the FK to users(id).
		// Wired to the "creator" edge below via Field("created_by").
		field.String("created_by").
			NotEmpty(),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Group.
func (Group) Edges() []ent.Edge {
	return []ent.Edge{
		// The user who created this group.
		// edge.From() points back to User.Edges → edge.To("created_groups")
		// Field("created_by") wires the FK column to this edge
		edge.From("creator", User.Type).
			Ref("created_groups").
			Field("created_by").
			Unique().   // Many groups → One creator
			Required(), // Every group must have a creator

		// Members of this group — through the GroupMember join entity.
		// We use GroupMember (not a raw many-to-many) because group_members
		// has extra columns: role and joined_at.
		edge.To("memberships", GroupMember.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),

		// Messages sent in this group.
		edge.To("messages", GroupMessage.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the Group.
func (Group) Indexes() []ent.Index {
	return []ent.Index{
		// Matches DB: CREATE INDEX groups_created_by_idx ON groups(created_by)
		index.Fields("created_by"),
	}
}

// Annotations tell Ent to use the exact existing table name.
func (Group) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "groups"},
	}
}
