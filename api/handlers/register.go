package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Register creates a new account.
func (h *Auth) Register(c *gin.Context) {
	var request credentialsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	_, err := h.auth.Register(c.Request.Context(), request.Username, request.Password)
	if err != nil {
		h.writeAuthError(c, err, true)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "sign up successful"})
}
