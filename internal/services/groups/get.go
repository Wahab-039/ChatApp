package groups

import (
	"context"

	"github.com/Wahab-039/ChatApp/internal/models"
)

// Get returns a group by ID if the requester is a member.
func (s *Service) Get(ctx context.Context, groupID, requesterID string) (models.GroupWithMembers, error) {
	isMember, err := s.groups.IsMember(ctx, groupID, requesterID)
	if err != nil {
		return models.GroupWithMembers{}, err
	}
	if !isMember {
		return models.GroupWithMembers{}, models.ErrNotGroupMember
	}

	group, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return models.GroupWithMembers{}, err
	}

	members, err := s.groups.ListMembers(ctx, groupID)
	if err != nil {
		return models.GroupWithMembers{}, err
	}

	return models.GroupWithMembers{
		Group:   group,
		Members: members,
	}, nil
}

// List returns all groups the requester is a member of.
func (s *Service) List(ctx context.Context, requesterID string) ([]models.Group, error) {
	return s.groups.ListByUserID(ctx, requesterID)
}
