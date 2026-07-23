package groups

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// Create creates a new group with the creator as admin.
func (s *Service) Create(ctx context.Context, creatorID, name string) (models.Group, error) {
	normalizedName, err := normalizeGroupName(name)
	if err != nil {
		return models.Group{}, err
	}

	group, err := s.groups.Create(ctx, normalizedName, creatorID)
	if err != nil {
		return models.Group{}, err
	}

	return group, nil
}

func normalizeGroupName(name string) (string, error) {
	normalized := strings.TrimSpace(name)
	if normalized == "" {
		return "", ErrGroupNameRequired
	}
	if utf8.RuneCountInString(normalized) > maxGroupNameLength {
		return "", ErrGroupNameTooLong
	}
	return normalized, nil
}
