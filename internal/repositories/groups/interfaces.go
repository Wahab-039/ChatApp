package groups

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// RepositoryInterface is the persistence contract for group operations.
type RepositoryInterface interface {
	Create(ctx context.Context, name, createdBy string) (models.Group, error)
	FindByID(ctx context.Context, id string) (models.Group, error)
	ListByUserID(ctx context.Context, userID string) ([]models.Group, error)
	AddMember(ctx context.Context, groupID, userID, role string) error
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
	ListMembers(ctx context.Context, groupID string) ([]models.User, error)
}
