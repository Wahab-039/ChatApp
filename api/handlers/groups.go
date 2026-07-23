package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/Wahab-039/ChatApp/api/middleware"
	"github.com/Wahab-039/ChatApp/internal/models"
	groupmessagesservice "github.com/Wahab-039/ChatApp/internal/services/groupmessages"
	groupsservice "github.com/Wahab-039/ChatApp/internal/services/groups"
	"github.com/gin-gonic/gin"
)

// GroupService defines group management operations used by HTTP handlers.
type GroupService interface {
	Create(ctx context.Context, creatorID, name string) (models.Group, error)
	Get(ctx context.Context, groupID, requesterID string) (models.GroupWithMembers, error)
	List(ctx context.Context, requesterID string) ([]models.Group, error)
	AddMember(ctx context.Context, groupID, adderID, username string) error
}

// GroupMessageService defines group message operations used by HTTP handlers.
type GroupMessageService interface {
	Send(ctx context.Context, senderID, groupID, body, clientMessageID string) (groupmessagesservice.SendResult, error)
	List(ctx context.Context, groupID, requesterID string, query groupmessagesservice.HistoryQuery) (groupmessagesservice.HistoryResult, error)
}

// Groups handles group-related HTTP requests.
type Groups struct {
	groups   GroupService
	messages GroupMessageService
}

// NewGroups creates a groups handler.
func NewGroups(groups GroupService, messages GroupMessageService) *Groups {
	return &Groups{groups: groups, messages: messages}
}

type createGroupRequest struct {
	Name string `json:"name"`
}

// Create creates a new group.
func (h *Groups) Create(c *gin.Context) {
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
func (h *Groups) Get(c *gin.Context) {
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
func (h *Groups) List(c *gin.Context) {
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
func (h *Groups) AddMember(c *gin.Context) {
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

type sendMessageRequest struct {
	Body            string `json:"body"`
	ClientMessageID string `json:"client_message_id"`
}

// SendMessage sends a message to a group.
func (h *Groups) SendMessage(c *gin.Context) {
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

	var request sendMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body and client_message_id are required"})
		return
	}

	result, err := h.messages.Send(
		c.Request.Context(),
		identity.UserID,
		groupID,
		request.Body,
		request.ClientMessageID,
	)
	if err != nil {
		h.writeMessageError(c, err, true)
		return
	}

	status := http.StatusCreated
	if !result.Created {
		status = http.StatusOK
	}

	c.JSON(status, gin.H{"message": result.Message})
}

// ListMessages returns paginated group message history.
func (h *Groups) ListMessages(c *gin.Context) {
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

	limit, err := groupmessagesservice.ParseLimit(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.messages.List(c.Request.Context(), groupID, identity.UserID, groupmessagesservice.HistoryQuery{
		BeforeID: c.Query("before"),
		AfterID:  c.Query("after"),
		Limit:    limit,
	})
	if err != nil {
		h.writeMessageError(c, err, false)
		return
	}

	response := gin.H{
		"messages": result.Messages,
	}
	if result.NextBefore != "" {
		response["next_before"] = result.NextBefore
	}
	if result.NextAfter != "" {
		response["next_after"] = result.NextAfter
	}

	c.JSON(http.StatusOK, response)
}

func (h *Groups) writeGroupError(c *gin.Context, err error) {
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

func (h *Groups) writeMessageError(c *gin.Context, err error, sending bool) {
	switch {
	case errors.Is(err, groupmessagesservice.ErrInvalidBody),
		errors.Is(err, groupmessagesservice.ErrInvalidClientMessageID),
		errors.Is(err, groupmessagesservice.ErrInvalidCursor),
		errors.Is(err, groupmessagesservice.ErrInvalidLimit):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, models.ErrNotGroupMember):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, models.ErrGroupNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case sending && errors.Is(err, models.ErrDuplicateClientMessage):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		if sending {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to send message"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to load messages"})
	}
}
