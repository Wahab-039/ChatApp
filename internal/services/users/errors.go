package users

import "errors"

// ErrSearchQueryRequired is returned when a user search request is empty.
var ErrSearchQueryRequired = errors.New("search query is required")
