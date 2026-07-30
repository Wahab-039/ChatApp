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

// GroupMember holds the schema definition for the GroupMember entity.
// The real table has a composite PK (group_id, user_id) with no surrogate id column.
// We name the schema field "id" but use StorageKey to map it to "group_id" column.
// The composite uniqueness is enforced by the unique index on (id, user_id).
type GroupMember struct {
	ent.Schema
}

// Fields of the GroupMember.
func (GroupMember) Fields() []ent.Field {
	return []ent.Field{
		// Ent's "id" field mapped to the "group_id" column via StorageKey.
		// This tells Ent to use group_id as the primary key column.
		field.String("id").
			StorageKey("group_id").
			NotEmpty().
			Immutable(),

		// user_id is a regular field that forms the second part of the composite PK.
		field.String("user_id").
			NotEmpty().
			Immutable(),

		// role: 'admin' or 'member' — matches DB CHECK constraint
		field.Enum("role").
			Values("admin", "member").
			Default("member"),

		field.Time("joined_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the GroupMember.
func (GroupMember) Edges() []ent.Edge {
	return []ent.Edge{
		// The group this membership belongs to.
		// The "id" field (mapped to group_id column) is automatically used as foreign key
		edge.From("group", Group.Type).
			Ref("memberships").
			Unique().
			Required().
			Immutable(),

		// The user who is the member.
		edge.From("user", User.Type).
			Ref("memberships").
			Field("user_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes of the GroupMember.
func (GroupMember) Indexes() []ent.Index {
	return []ent.Index{
		// Composite uniqueness — mirrors the real DB PRIMARY KEY (group_id, user_id).
		// In schema we use "id" which maps to group_id column.
		index.Fields("id", "user_id").Unique(),

		index.Fields("user_id"),
		index.Fields("id"),
	}
}

// Annotations tell Ent to use the exact existing table name.
func (GroupMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "group_members"},
	}
}
