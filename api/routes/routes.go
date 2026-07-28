package routes

import (
	authhandler "github.com/Wahab-039/ChatApp/api/handlers/auth"
	groupmessageshandler "github.com/Wahab-039/ChatApp/api/handlers/groupmessages"
	groupshandler "github.com/Wahab-039/ChatApp/api/handlers/groups"
	healthhandler "github.com/Wahab-039/ChatApp/api/handlers/health"
	messageshandler "github.com/Wahab-039/ChatApp/api/handlers/messages"
	mqttdevhandler "github.com/Wahab-039/ChatApp/api/handlers/mqttdev"
	usershandler "github.com/Wahab-039/ChatApp/api/handlers/users"
	"github.com/gin-gonic/gin"
)

// Register wires HTTP routes to their fully constructed handlers.
func Register(
	router *gin.Engine,
	health healthhandler.HandlerInterface,
	authHandler authhandler.HandlerInterface,
	usersHandler usershandler.HandlerInterface,
	messagesHandler messageshandler.HandlerInterface,
	groupsHandler groupshandler.HandlerInterface,
	groupMessagesHandler groupmessageshandler.HandlerInterface,
	requireAuth gin.HandlerFunc,
	mqttDev *mqttdevhandler.Handler,
) {
	router.GET("/health", health.Check)

	api := router.Group("/api/v1")
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	protected := api.Group("")
	protected.Use(requireAuth)
	protected.GET("/me", usersHandler.Me)
	protected.GET("/users", usersHandler.SearchUsers)
	protected.POST("/messages/direct", messagesHandler.SendDirect)
	protected.GET("/messages/direct", messagesHandler.ListDirect)

	protected.POST("/groups", groupsHandler.Create)
	protected.GET("/groups", groupsHandler.List)
	protected.GET("/groups/:id", groupsHandler.Get)
	protected.POST("/groups/:id/members", groupsHandler.AddMember)
	protected.POST("/groups/:id/messages", groupMessagesHandler.SendMessage)
	protected.GET("/groups/:id/messages", groupMessagesHandler.ListMessages)

	if mqttDev != nil {
		protected.POST("/dev/mqtt/ping", mqttDev.Ping)
	}
}
