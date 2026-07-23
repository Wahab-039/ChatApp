package groups

import (
	"context"
	"errors"
	"strings"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// AddMember adds a user to a group.
func (s *Service) AddMember(ctx context.Context, groupID, adderID, username string) error {
	// Verify adder is a member
	isMember, err := s.groups.IsMember(ctx, groupID, adderID)
	if err != nil {
		return err
	}
	if !isMember {
		return models.ErrNotGroupMember
	}

	// Look up user to add
	normalizedUsername := strings.ToLower(strings.TrimSpace(username))
	if normalizedUsername == "" {
		return ErrUsernameRequired
	}

	user, err := s.users.FindByUsername(ctx, normalizedUsername)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Add as member
	if err := s.groups.AddMember(ctx, groupID, user.ID, "member"); err != nil {
		return err
	}

	return nil
}
