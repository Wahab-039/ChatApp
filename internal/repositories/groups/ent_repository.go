package groups

import (
	"context"
	"fmt"

	"github.com/Wahab-039/ChatApp/ent"
	"github.com/Wahab-039/ChatApp/ent/group"
	"github.com/Wahab-039/ChatApp/ent/groupmember"
	"github.com/Wahab-039/ChatApp/ent/user"
	"github.com/Wahab-039/ChatApp/internal/models"
)

// EntRepository stores groups using the Ent ORM client.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository creates a groups repository backed by the Ent client.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

// Compile-time check: EntRepository must satisfy RepositoryInterface.
var _ RepositoryInterface = (*EntRepository)(nil)

// toGroup converts an Ent Group entity to the application model.
func toGroup(e *ent.Group) models.Group {
	return models.Group{
		ID:        e.ID,
		Name:      e.Name,
		CreatedBy: e.CreatedBy,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

// toUser converts an Ent User entity to the application model.
func toUser(e *ent.User) models.User {
	return models.User{
		ID:        e.ID,
		Username:  e.Username,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

// Create inserts a new group and adds the creator as admin in a single transaction.
func (r *EntRepository) Create(ctx context.Context, name, createdBy string) (models.Group, error) {
	// Use Ent transaction to ensure both operations succeed or fail together.
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return models.Group{}, fmt.Errorf("begin transaction: %w", err)
	}

	// Rollback on error, commit explicitly at the end.
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	// Create the group.
	g, err := tx.Group.Create().
		SetName(name).
		SetCreatedBy(createdBy).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return models.Group{}, fmt.Errorf("create group: %w", err)
	}

	// Add the creator as admin member.
	_, err = tx.GroupMember.Create().
		SetID(g.ID). // ID maps to group_id column via StorageKey
		SetUserID(createdBy).
		SetRole(groupmember.RoleAdmin).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return models.Group{}, fmt.Errorf("add creator as admin: %w", err)
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		return models.Group{}, fmt.Errorf("commit transaction: %w", err)
	}

	return toGroup(g), nil
}

// FindByID returns a group by its id.
func (r *EntRepository) FindByID(ctx context.Context, id string) (models.Group, error) {
	g, err := r.client.Group.Query().
		Where(group.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return models.Group{}, models.ErrGroupNotFound
		}
		return models.Group{}, fmt.Errorf("find group by id: %w", err)
	}

	return toGroup(g), nil
}

// ListByUserID returns all groups the user belongs to, newest updated first.
func (r *EntRepository) ListByUserID(ctx context.Context, userID string) ([]models.Group, error) {
	entities, err := r.client.Group.Query().
		Where(
			group.HasMembershipsWith(
				groupmember.UserID(userID),
			),
		).
		Order(ent.Desc(group.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list groups by user: %w", err)
	}

	groups := make([]models.Group, len(entities))
	for i, g := range entities {
		groups[i] = toGroup(g)
	}

	return groups, nil
}

// AddMember inserts a row into group_members with the given role.
func (r *EntRepository) AddMember(ctx context.Context, groupID, userID, role string) error {
	// Validate role before attempting insert.
	if err := groupmember.RoleValidator(groupmember.Role(role)); err != nil {
		return fmt.Errorf("invalid role %q: %w", role, err)
	}

	// Use Ent's Create builder now that StorageKey fixes the id column issue.
	_, err := r.client.GroupMember.Create().
		SetID(groupID). // ID maps to group_id column via StorageKey
		SetUserID(userID).
		SetRole(groupmember.Role(role)).
		Save(ctx)

	if err != nil {
		// Check for unique constraint violation (user already in group).
		if ent.IsConstraintError(err) {
			return models.ErrAlreadyGroupMember
		}
		return fmt.Errorf("add group member: %w", err)
	}

	return nil
}

// IsMember checks whether userID has a membership row for groupID.
func (r *EntRepository) IsMember(ctx context.Context, groupID, userID string) (bool, error) {
	count, err := r.client.GroupMember.Query().
		Where(
			groupmember.ID(groupID), // ID now maps to group_id column
			groupmember.UserID(userID),
		).
		Count(ctx)
	if err != nil {
		return false, fmt.Errorf("check group membership: %w", err)
	}

	return count > 0, nil
}

// ListMembers returns all users who are members of groupID.
func (r *EntRepository) ListMembers(ctx context.Context, groupID string) ([]models.User, error) {
	entities, err := r.client.User.Query().
		Where(
			user.HasMembershipsWith(
				groupmember.ID(groupID), // ID now maps to group_id column
			),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list group members: %w", err)
	}

	members := make([]models.User, len(entities))
	for i, e := range entities {
		members[i] = toUser(e)
	}

	return members, nil
}
