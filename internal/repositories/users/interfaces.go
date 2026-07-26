package users

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// RepositoryInterface is the persistence contract for user operations.
type RepositoryInterface interface {
	Create(ctx context.Context, username, passwordHash string) (models.User, error)
	FindByID(ctx context.Context, id string) (models.User, error)
	FindByUsername(ctx context.Context, username string) (models.Credentials, error)
	SearchByUsername(ctx context.Context, query, excludedUserID string, limit int) ([]models.User, error)
}
