package groups

import (
	"errors"
	"net/http"

	"github.com/Wahab-039/ChatApp/api/middleware"
	"github.com/Wahab-039/ChatApp/internal/models"
	groupsservice "github.com/Wahab-039/ChatApp/internal/services/groups"
	"github.com/gin-gonic/gin"
)

type createGroupRequest struct {
	Name string `json:"name"`
}

// Create creates a new group.
func (h *Handler) Create(c *gin.Context) {
	identity, ok := middleware.IdentityFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication is required"})
		return
	}

	var request createGroupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	group, err := h.groups.Create(c.Request.Context(), identity.UserID, request.Name)
	if err != nil {
		h.writeGroupError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"group": group})
}

// Get returns a group with its members.
func (h *Handler) Get(c *gin.Context) {
	identity, ok := middleware.IdentityFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication is required"})
		return
	}

	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group id is required"})
		return
	}

	group, err := h.groups.Get(c.Request.Context(), groupID, identity.UserID)
	if err != nil {
		h.writeGroupError(c, err)
		return
	}

	c.JSON(http.StatusOK, group)
}

// List returns all groups the user is a member of.
func (h *Handler) List(c *gin.Context) {
	identity, ok := middleware.IdentityFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication is required"})
		return
	}

	groups, err := h.groups.List(c.Request.Context(), identity.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to list groups"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

type addMemberRequest struct {
	Username string `json:"username"`
}

// AddMember adds a user to a group.
func (h *Handler) AddMember(c *gin.Context) {
	identity, ok := middleware.IdentityFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication is required"})
		return
	}

	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group id is required"})
		return
	}

	var request addMemberRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	err := h.groups.AddMember(c.Request.Context(), groupID, identity.UserID, request.Username)
	if err != nil {
		h.writeGroupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member added successfully"})
}

func (h *Handler) writeGroupError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, groupsservice.ErrGroupNameRequired),
		errors.Is(err, groupsservice.ErrGroupNameTooLong),
		errors.Is(err, groupsservice.ErrUsernameRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, models.ErrGroupNotFound),
		errors.Is(err, groupsservice.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, models.ErrNotGroupMember):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, models.ErrAlreadyGroupMember):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to process request"})
	}
}
