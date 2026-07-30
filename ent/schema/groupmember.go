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
// This represents the group_members table, which is NOT a simple junction table —
// it has extra columns (role, joined_at), so it needs its own full schema.
type GroupMember struct {
	ent.Schema
}

// Fields of the GroupMember.
func (GroupMember) Fields() []ent.Field {
	return []ent.Field{
		// group_id and user_id are foreign keys wired to edges below.
		// Together they form the composite PK in the real DB table.
		field.String("group_id").
			NotEmpty(),

		field.String("user_id").
			NotEmpty(),

		// role: 'admin' or 'member' — matches DB CHECK constraint
		field.Enum("role").
			Values("admin", "member").
			Default("member"),

		field.Time("joined_at").
			Default(time.Now).
			Immutable(), // Membership timestamp never changes
	}
}

// Edges of the GroupMember.
func (GroupMember) Edges() []ent.Edge {
	return []ent.Edge{
		// The group this membership belongs to.
		// Ref("memberships") links back to Group.Edges → edge.To("memberships")
		edge.From("group", Group.Type).
			Ref("memberships").
			Field("group_id").
			Unique().   // Many memberships → One group
			Required(),

		// The user who is the member.
		// Ref("memberships") links back to User.Edges → edge.To("memberships")
		edge.From("user", User.Type).
			Ref("memberships").
			Field("user_id").
			Unique().   // Many memberships → One user
			Required(),
	}
}

// Indexes of the GroupMember.
func (GroupMember) Indexes() []ent.Index {
	return []ent.Index{
		// The real DB primary key is (group_id, user_id).
		// Declaring it as a unique index here enforces the same constraint.
		index.Fields("group_id", "user_id").Unique(),

		// Matches DB: CREATE INDEX group_members_user_idx ON group_members(user_id)
		index.Fields("user_id"),

		// Matches DB: CREATE INDEX group_members_group_idx ON group_members(group_id)
		index.Fields("group_id"),
	}
}

// Annotations tell Ent to use the exact existing table name.
func (GroupMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "group_members"},
	}
}
