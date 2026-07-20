package routes

import (
	"github.com/Wahab-039/ChatApp/api/handlers"
	"github.com/gin-gonic/gin"
)

// Register wires HTTP routes to their fully constructed handlers.
func Register(
	router *gin.Engine,
	health *handlers.Health,
	authHandler *handlers.Auth,
	limitLogin gin.HandlerFunc,
	requireAuth gin.HandlerFunc,
) {
	router.GET("/health", health.Check)

	api := router.Group("/api/v1")
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", limitLogin, authHandler.Login)

	protected := api.Group("")
	protected.Use(requireAuth)
	protected.GET("/me", authHandler.Me)
	protected.GET("/users", authHandler.SearchUsers)
}
