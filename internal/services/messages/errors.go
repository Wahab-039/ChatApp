package messages

import "errors"

var (
	// ErrRecipientRequired is returned when recipient_username is missing.
	ErrRecipientRequired = errors.New("recipient_username is required")
	// ErrPeerRequired is returned when the history "with" username is missing.
	ErrPeerRequired = errors.New("with username is required")
	// ErrRecipientNotFound is returned when the recipient username does not exist.
	ErrRecipientNotFound = errors.New("recipient not found")
	// ErrCannotMessageSelf is returned when sender and recipient are the same user.
	ErrCannotMessageSelf = errors.New("cannot message yourself")
	// ErrInvalidBody is returned when the message body is empty or too long.
	ErrInvalidBody = errors.New("body must be between 1 and 4000 characters")
	// ErrInvalidClientMessageID is returned when client_message_id is missing or too long.
	ErrInvalidClientMessageID = errors.New("client_message_id must be between 1 and 128 characters")
	// ErrInvalidCursor is returned when before/after point to an invalid message.
	ErrInvalidCursor = errors.New("invalid before/after cursor")
	// ErrInvalidLimit is returned when limit is not in the allowed range.
	ErrInvalidLimit = errors.New("limit must be between 1 and 100")
)
