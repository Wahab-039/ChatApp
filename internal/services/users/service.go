// Package users contains user-profile use cases.
package users

// Service contains dependencies shared by user-management use cases.
type Service struct {
	repository UserRepository
}

// NewService creates a user-management service with an explicit repository dependency.
func NewService(repository UserRepository) *Service {
	return &Service{repository: repository}
}
