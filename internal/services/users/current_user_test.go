package users

import (
	"context"
	"testing"

	"github.com/Wahab-039/ChatApp/internal/models"
)

type fakeProfileRepository struct {
	user         models.User
	err          error
	searchUsers  []models.User
	searchErr    error
	searchQuery  string
	excludedUser string
	searchLimit  int
}

func (r fakeProfileRepository) FindByID(_ context.Context, _ string) (models.User, error) {
	return r.user, r.err
}

func (r *fakeProfileRepository) SearchByUsername(
	_ context.Context,
	query, excludedUserID string,
	limit int,
) ([]models.User, error) {
	r.searchQuery = query
	r.excludedUser = excludedUserID
	r.searchLimit = limit
	return r.searchUsers, r.searchErr
}

func TestCurrentUserReturnsRepositoryProfile(t *testing.T) {
	want := models.User{ID: "user-1", Username: "wahab"}
	service := NewService(&fakeProfileRepository{user: want})

	got, err := service.CurrentUser(context.Background(), want.ID)
	if err != nil {
		t.Fatalf("CurrentUser() error = %v", err)
	}
	if got != want {
		t.Fatalf("CurrentUser() = %#v, want %#v", got, want)
	}
}

var _ Repository = (*fakeProfileRepository)(nil)
