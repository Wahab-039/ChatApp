package users

import (
	"context"
	"strings"

	"github.com/Wahab-039/ChatApp/internal/models"
)

const searchResultLimit = 20

// Search returns users whose usernames begin with query, excluding the requester.
func (s *Service) Search(ctx context.Context, query, requesterID string) ([]models.User, error) {
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	if normalizedQuery == "" {
		return nil, ErrSearchQueryRequired
	}

	return s.repository.SearchByUsername(ctx, normalizedQuery, requesterID, searchResultLimit)
}
