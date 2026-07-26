package messages

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// RepositoryInterface is the persistence contract for direct message operations.
type RepositoryInterface interface {
	Create(ctx context.Context, senderID, recipientID, body, clientMessageID string) (models.DirectMessage, error)
	FindByID(ctx context.Context, id string) (models.DirectMessage, error)
	FindBySenderAndClientMessageID(ctx context.Context, senderID, clientMessageID string) (models.DirectMessage, error)
	ListConversation(ctx context.Context, userID, peerID string, before, after *models.DirectMessage, limit int) ([]models.DirectMessage, error)
}
