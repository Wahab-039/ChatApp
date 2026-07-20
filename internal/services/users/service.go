// Package users contains user-profile use cases.
package users

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// Repository is the persistence contract required by user-management use cases.
type Repository interface {
	FindByID(ctx context.Context, id string) (models.User, error)
	SearchByUsername(ctx context.Context, query, excludedUserID string, limit int) ([]models.User, error)
}

// Service contains dependencies shared by user-management use cases.
type Service struct {
	repository Repository
}

// NewService creates a user-management service with an explicit repository dependency.
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}
