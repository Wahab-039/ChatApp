package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Login validates account credentials and returns an access token.
func (h *Auth) Login(c *gin.Context) {
	var request credentialsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	result, err := h.auth.Login(c.Request.Context(), request.Username, request.Password)
	if err != nil {
		h.writeAuthError(c, err, false)
		return
	}

	c.JSON(http.StatusOK, struct {
		Message     string `json:"message"`
		AccessToken string `json:"access_token"`
	}{
		Message:     "login successful",
		AccessToken: result.AccessToken,
	})
}
