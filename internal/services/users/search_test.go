package users

import (
	"context"
	"errors"
	"testing"

	"github.com/Wahab-039/ChatApp/internal/models"
)

func TestSearchNormalizesQueryAndExcludesRequester(t *testing.T) {
	repository := &fakeProfileRepository{
		searchUsers: []models.User{{ID: "user-2", Username: "wahab_2"}},
	}
	service := NewService(repository)

	users, err := service.Search(context.Background(), "  WAH  ", "user-1")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(users) != 1 || users[0].Username != "wahab_2" {
		t.Fatalf("Search() users = %#v", users)
	}
	if repository.searchQuery != "wah" {
		t.Fatalf("search query = %q, want %q", repository.searchQuery, "wah")
	}
	if repository.excludedUser != "user-1" {
		t.Fatalf("excluded user = %q, want %q", repository.excludedUser, "user-1")
	}
	if repository.searchLimit != searchResultLimit {
		t.Fatalf("search limit = %d, want %d", repository.searchLimit, searchResultLimit)
	}
}

func TestSearchRejectsEmptyQuery(t *testing.T) {
	service := NewService(&fakeProfileRepository{})

	_, err := service.Search(context.Background(), "   ", "user-1")
	if !errors.Is(err, ErrSearchQueryRequired) {
		t.Fatalf("Search() error = %v, want ErrSearchQueryRequired", err)
	}
}
