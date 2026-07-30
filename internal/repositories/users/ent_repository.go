package users

import (
	"context"
	"fmt"

	"github.com/Wahab-039/ChatApp/ent"
	"github.com/Wahab-039/ChatApp/ent/user"
	"github.com/Wahab-039/ChatApp/internal/models"
)

// EntRepository stores users using the Ent ORM client.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository creates a user repository backed by the Ent client.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

// Compile-time check: EntRepository must satisfy RepositoryInterface.
var _ RepositoryInterface = (*EntRepository)(nil)

// toUser converts an Ent User entity to the application model.
// This is a pure data mapping — no business logic here.
func toUser(e *ent.User) models.User {
	return models.User{
		ID:        e.ID,
		Username:  e.Username,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

// toCredentials converts an Ent User entity to Credentials (includes password hash).
func toCredentials(e *ent.User) models.Credentials {
	return models.Credentials{
		User:         toUser(e),
		PasswordHash: e.PasswordHash,
	}
}

// Create inserts a new user row and returns the created user.
// The DB generates the id and timestamps via DEFAULT expressions.
func (r *EntRepository) Create(ctx context.Context, username, passwordHash string) (models.User, error) {
	e, err := r.client.User.Create().
		SetUsername(username).
		SetPasswordHash(passwordHash).
		Save(ctx)
	if err != nil {
		// ent.IsConstraintError catches UNIQUE violations (username taken)
		if ent.IsConstraintError(err) {
			return models.User{}, models.ErrUsernameTaken
		}
		return models.User{}, fmt.Errorf("create user: %w", err)
	}

	return toUser(e), nil
}

// FindByID returns the public profile for the given id.
func (r *EntRepository) FindByID(ctx context.Context, id string) (models.User, error) {
	e, err := r.client.User.Query().
		Where(user.ID(id)). // generated predicate: WHERE id = $1
		Only(ctx)           // expects exactly one row; errors if 0 or >1
	if err != nil {
		if ent.IsNotFound(err) {
			return models.User{}, models.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("find user by id: %w", err)
	}

	return toUser(e), nil
}

// FindByUsername returns credentials for authenticating the given username.
func (r *EntRepository) FindByUsername(ctx context.Context, username string) (models.Credentials, error) {
	e, err := r.client.User.Query().
		Where(user.Username(username)). // WHERE username = $1
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return models.Credentials{}, models.ErrUserNotFound
		}
		return models.Credentials{}, fmt.Errorf("find user by username: %w", err)
	}

	return toCredentials(e), nil
}

// SearchByUsername returns users whose usernames begin with query, excluding excludedUserID.
// Results are ordered alphabetically and capped at limit.
func (r *EntRepository) SearchByUsername(
	ctx context.Context,
	query, excludedUserID string,
	limit int,
) ([]models.User, error) {
	entities, err := r.client.User.Query().
		Where(
			user.UsernameHasPrefix(query),    // WHERE username LIKE $1 || '%'
			user.IDNEQ(excludedUserID),       // AND id <> $2
		).
		Order(ent.Asc(user.FieldUsername)). // ORDER BY username ASC
		Limit(limit).
		All(ctx) // returns a slice (vs Only which expects one)
	if err != nil {
		return nil, fmt.Errorf("search users by username: %w", err)
	}

	users := make([]models.User, len(entities))
	for i, e := range entities {
		users[i] = toUser(e)
	}

	return users, nil
}
