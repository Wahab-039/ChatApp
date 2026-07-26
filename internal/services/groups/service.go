// Package groups contains group management use cases.
package groups

import (
	"github.com/Wahab-039/ChatApp/internal/services/users"
)

const maxGroupNameLength = 100

// Service handles group management use cases.
type Service struct {
	groups GroupRepository
	users  users.UserRepository
}

// NewService creates a groups service.
func NewService(groups GroupRepository, users users.UserRepository) *Service {
	return &Service{groups: groups, users: users}
}
