package mqtt

import "context"

// InboxPublisher delivers events to recipient inboxes via MQTT.
type InboxPublisher interface {
	PublishToUserInbox(ctx context.Context, userID string, event Event) error
}
