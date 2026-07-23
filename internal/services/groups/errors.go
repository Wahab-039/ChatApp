package groups

import "errors"

var (
	// ErrGroupNameRequired is returned when group name is empty.
	ErrGroupNameRequired = errors.New("group name is required")
	// ErrGroupNameTooLong is returned when group name exceeds 100 characters.
	ErrGroupNameTooLong = errors.New("group name must be 100 characters or less")
	// ErrUsernameRequired is returned when username to add is empty.
	ErrUsernameRequired = errors.New("username is required")
	// ErrUserNotFound is returned when the user to add does not exist.
	ErrUserNotFound = errors.New("user not found")
)
