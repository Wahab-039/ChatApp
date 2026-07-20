package handlers

import (
	"errors"
	"net/http"

	"github.com/Wahab-039/ChatApp/internal/models"
	authservice "github.com/Wahab-039/ChatApp/internal/services/auth"
	"github.com/gin-gonic/gin"
)

func (h *Auth) writeAuthError(c *gin.Context, err error, registration bool) {
	switch {
	case errors.Is(err, models.ErrUsernameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "username is already taken"})
	case registration && (errors.Is(err, authservice.ErrInvalidUsername) || errors.Is(err, authservice.ErrInvalidPassword)):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, authservice.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to process authentication request"})
	}
}
