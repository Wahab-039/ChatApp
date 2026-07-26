package groupmessages

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// RepositoryInterface is the persistence contract for group message operations.
type RepositoryInterface interface {
	Create(ctx context.Context, groupID, senderID, body, clientMessageID string) (models.GroupMessage, error)
	FindByID(ctx context.Context, id string) (models.GroupMessage, error)
	FindBySenderAndClientMessageID(ctx context.Context, senderID, clientMessageID string) (models.GroupMessage, error)
	ListByGroup(ctx context.Context, groupID string, before, after *models.GroupMessage, limit int) ([]models.GroupMessage, error)
}
