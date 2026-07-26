// Package users contains user-profile use cases.
package users

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
	userrepository "github.com/Wahab-039/ChatApp/internal/repositories/users"
)

// ServiceInterface defines user-profile use cases.
type ServiceInterface interface {
	CurrentUser(ctx context.Context, id string) (models.User, error)
	Search(ctx context.Context, query, requesterID string) ([]models.User, error)
}

// Service contains dependencies shared by user-management use cases.
type Service struct {
	repository userrepository.RepositoryInterface
}

// NewService creates a user-management service with an explicit repository dependency.
func NewService(repository userrepository.RepositoryInterface) *Service {
	return &Service{repository: repository}
}

var _ ServiceInterface = (*Service)(nil)
