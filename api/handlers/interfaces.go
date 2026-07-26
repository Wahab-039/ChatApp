package handlers

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
	appmqtt "github.com/Wahab-039/ChatApp/internal/mqtt"
	authservice "github.com/Wahab-039/ChatApp/internal/services/auth"
	groupmessagesservice "github.com/Wahab-039/ChatApp/internal/services/groupmessages"
	messagesservice "github.com/Wahab-039/ChatApp/internal/services/messages"
)

// AuthService defines the authentication operations used by HTTP handlers.
type AuthService interface {
	Register(ctx context.Context, username, password string) (models.User, error)
	Login(ctx context.Context, username, password string) (authservice.Result, error)
}

// UserService defines the user-profile operations used by HTTP handlers.
type UserService interface {
	CurrentUser(ctx context.Context, id string) (models.User, error)
	Search(ctx context.Context, query, requesterID string) ([]models.User, error)
}

// DirectMessageService defines message operations used by HTTP handlers.
type DirectMessageService interface {
	SendDirect(ctx context.Context, senderID, recipientUsername, body, clientMessageID string) (messagesservice.SendResult, error)
	ListDirect(ctx context.Context, requesterID string, query messagesservice.HistoryQuery) (messagesservice.HistoryResult, error)
}

// GroupService defines group management operations used by HTTP handlers.
type GroupService interface {
	Create(ctx context.Context, creatorID, name string) (models.Group, error)
	Get(ctx context.Context, groupID, requesterID string) (models.GroupWithMembers, error)
	List(ctx context.Context, requesterID string) ([]models.Group, error)
	AddMember(ctx context.Context, groupID, adderID, username string) error
}

// GroupMessageService defines group message operations used by HTTP handlers.
type GroupMessageService interface {
	Send(ctx context.Context, senderID, groupID, body, clientMessageID string) (groupmessagesservice.SendResult, error)
	List(ctx context.Context, groupID, requesterID string, query groupmessagesservice.HistoryQuery) (groupmessagesservice.HistoryResult, error)
}

// InboxPublisher is the minimal MQTT contract used by the dev ping endpoint.
type InboxPublisher interface {
	PublishToUserInbox(ctx context.Context, userID string, event appmqtt.Event) error
}

// DatabaseHealthChecker is the minimal database contract required by the health endpoint.
type DatabaseHealthChecker interface {
	Ping(ctx context.Context) error
}
