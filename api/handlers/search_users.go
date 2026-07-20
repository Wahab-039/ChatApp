package handlers

import (
	"errors"
	"net/http"

	"github.com/Wahab-039/ChatApp/api/middleware"
	userservice "github.com/Wahab-039/ChatApp/internal/services/users"
	"github.com/gin-gonic/gin"
)

type userSearchResponse struct {
	Username string `json:"username"`
}

// SearchUsers returns users whose usernames begin with the query parameter.
func (h *Auth) SearchUsers(c *gin.Context) {
	identity, ok := middleware.IdentityFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authenticated identity is unavailable"})
		return
	}

	users, err := h.users.Search(c.Request.Context(), c.Query("query"), identity.UserID)
	if errors.Is(err, userservice.ErrSearchQueryRequired) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to search users"})
		return
	}

	response := make([]userSearchResponse, len(users))
	for index, user := range users {
		response[index] = userSearchResponse{Username: user.Username}
	}

	c.JSON(http.StatusOK, gin.H{"users": response})
}
