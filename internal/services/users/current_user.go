package users

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// CurrentUser returns the current profile for id.
func (s *Service) CurrentUser(ctx context.Context, id string) (models.User, error) {
	return s.repository.FindByID(ctx, id)
}
