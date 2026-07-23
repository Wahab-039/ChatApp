package models

import (
	"errors"
	"time"
)

var (
	// ErrGroupMessageNotFound is returned when no group message matches a query.
	ErrGroupMessageNotFound = errors.New("group message not found")
)

// GroupMessage is a persisted group chat message.
type GroupMessage struct {
	ID              string    `json:"id"`
	GroupID         string    `json:"group_id"`
	SenderID        string    `json:"sender_id"`
	Body            string    `json:"body"`
	ClientMessageID string    `json:"client_message_id"`
	CreatedAt       time.Time `json:"created_at"`
}
