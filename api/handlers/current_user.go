package handlers

import (
	"errors"
	"net/http"

	"github.com/Wahab-039/ChatApp/api/middleware"
	"github.com/Wahab-039/ChatApp/internal/models"
	"github.com/gin-gonic/gin"
)

// Me returns the authenticated user's current profile.
func (h *Auth) Me(c *gin.Context) {
	identity, ok := middleware.IdentityFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authenticated identity is unavailable"})
		return
	}

	user, err := h.users.CurrentUser(c.Request.Context(), identity.UserID)
	if errors.Is(err, models.ErrUserNotFound) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user account no longer exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to retrieve user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
