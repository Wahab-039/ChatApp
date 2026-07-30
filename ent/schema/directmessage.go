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

// DirectMessage holds the schema definition for the DirectMessage entity.
type DirectMessage struct {
	ent.Schema
}

// Fields of the DirectMessage.
func (DirectMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Immutable(),

		// sender_id and recipient_id are stored as plain strings (matching models.DirectMessage)
		// They also serve as foreign keys — wired up in Edges() below
		field.String("sender_id").
			NotEmpty(),

		field.String("recipient_id").
			NotEmpty(),

		// body maps to models.DirectMessage.Body
		field.String("body").
			NotEmpty(),

		// client_message_id: client-generated deduplication ID
		// Unique per sender — enforced by DB unique constraint + index below
		field.String("client_message_id").
			MaxLen(128).
			NotEmpty(),

		field.Time("created_at").
			Default(time.Now).
			Immutable(), // Messages are never edited
	}
}

// Edges of the DirectMessage.
// edge.From() points BACK to User, referencing the edges defined on User.
func (DirectMessage) Edges() []ent.Edge {
	return []ent.Edge{
		// This message was sent BY a User.
		// Ref("sent_messages") links back to User.Edges → edge.To("sent_messages")
		// Field("sender_id") tells Ent: store the FK in the sender_id column
		edge.From("sender", User.Type).
			Ref("sent_messages").
			Field("sender_id").
			Unique().   // Many messages → One sender
			Required(), // Every message must have a sender

		// This message was sent TO a User.
		// Ref("received_messages") links back to User.Edges → edge.To("received_messages")
		edge.From("recipient", User.Type).
			Ref("received_messages").
			Field("recipient_id").
			Unique().
			Required(),
	}
}

// Indexes of the DirectMessage.
func (DirectMessage) Indexes() []ent.Index {
	return []ent.Index{
		// Matches DB: UNIQUE (sender_id, client_message_id)
		// Prevents the same client from sending the same message twice
		index.Fields("sender_id", "client_message_id").Unique(),

		// Matches DB index for conversation queries (both directions)
		index.Fields("sender_id", "recipient_id", "created_at"),

		// Matches DB index for recipient inbox queries
		index.Fields("recipient_id", "created_at"),
	}
}

// Annotations tell Ent to use the exact existing table name.
func (DirectMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "direct_messages"},
	}
}
