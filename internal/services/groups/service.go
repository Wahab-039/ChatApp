// Package groups contains group management use cases.
package groups

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
)

const maxGroupNameLength = 100

// GroupRepository is the persistence contract for groups.
type GroupRepository interface {
	Create(ctx context.Context, name, createdBy string) (models.Group, error)
	FindByID(ctx context.Context, id string) (models.Group, error)
	ListByUserID(ctx context.Context, userID string) ([]models.Group, error)
	AddMember(ctx context.Context, groupID, userID, role string) error
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
	ListMembers(ctx context.Context, groupID string) ([]models.User, error)
}

// UserRepository looks up users for validation.
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (models.Credentials, error)
}

// Service handles group management use cases.
type Service struct {
	groups GroupRepository
	users  UserRepository
}

// NewService creates a groups service.
func NewService(groups GroupRepository, users UserRepository) *Service {
	return &Service{groups: groups, users: users}
}
