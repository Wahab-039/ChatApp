package models

import (
	"errors"
	"time"
)

var (
	// ErrDuplicateClientMessage is returned when a sender reuses a client_message_id.
	ErrDuplicateClientMessage = errors.New("client message id already used")
	// ErrMessageNotFound is returned when no direct message matches a query.
	ErrMessageNotFound = errors.New("message not found")
)

// DirectMessage is a persisted one-to-one chat message.
type DirectMessage struct {
	ID              string    `json:"id"`
	SenderID        string    `json:"sender_id"`
	RecipientID     string    `json:"recipient_id"`
	Body            string    `json:"body"`
	ClientMessageID string    `json:"client_message_id"`
	CreatedAt       time.Time `json:"created_at"`
}
