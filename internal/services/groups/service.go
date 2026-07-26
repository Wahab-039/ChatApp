// Package groups contains group management use cases.
package groups

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
	grouprepository "github.com/Wahab-039/ChatApp/internal/repositories/groups"
	userrepository "github.com/Wahab-039/ChatApp/internal/repositories/users"
)

const maxGroupNameLength = 100

// ServiceInterface defines group management use cases.
type ServiceInterface interface {
	Create(ctx context.Context, creatorID, name string) (models.Group, error)
	Get(ctx context.Context, groupID, requesterID string) (models.GroupWithMembers, error)
	List(ctx context.Context, requesterID string) ([]models.Group, error)
	AddMember(ctx context.Context, groupID, adderID, username string) error
}

// Service handles group management use cases.
type Service struct {
	groups grouprepository.RepositoryInterface
	users  userrepository.RepositoryInterface
}

// NewService creates a groups service.
func NewService(groups grouprepository.RepositoryInterface, users userrepository.RepositoryInterface) *Service {
	return &Service{groups: groups, users: users}
}

var _ ServiceInterface = (*Service)(nil)
