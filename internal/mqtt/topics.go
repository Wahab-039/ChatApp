package mqtt

import "fmt"

// UserInboxTopic returns the private inbox topic for a user UUID.
func UserInboxTopic(userID string) string {
	return fmt.Sprintf("chat/user/%s/inbox", userID)
}
