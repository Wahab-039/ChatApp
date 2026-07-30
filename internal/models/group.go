package models

import (
	"errors"
	"time"
)

var (
	// ErrGroupNotFound is returned when no group matches a query.
	ErrGroupNotFound = errors.New("group not found")
	// ErrNotGroupMember is returned when a user tries to access a group they're not in.
	ErrNotGroupMember = errors.New("you are not a member of this group")
	// ErrAlreadyGroupMember is returned when trying to add an existing member.
	ErrAlreadyGroupMember = errors.New("user is already a member")
)

// Group represents a multi-user chat group.
type Group struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GroupWithMembers includes the group and its member list.
type GroupWithMembers struct {
	Group
	Members []User `json:"members"`
}
