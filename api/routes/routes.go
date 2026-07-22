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
	messagesHandler *handlers.Messages,
	limitLogin gin.HandlerFunc,
	requireAuth gin.HandlerFunc,
	mqttDev *handlers.MQTTDev,
) {
	router.GET("/health", health.Check)

	api := router.Group("/api/v1")
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", limitLogin, authHandler.Login)

	protected := api.Group("")
	protected.Use(requireAuth)
	protected.GET("/me", authHandler.Me)
	protected.GET("/users", authHandler.SearchUsers)
	protected.POST("/messages/direct", messagesHandler.SendDirect)
	protected.GET("/messages/direct", messagesHandler.ListDirect)

	if mqttDev != nil {
		protected.POST("/dev/mqtt/ping", mqttDev.Ping)
	}
}
