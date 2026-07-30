package app

import (
	"github.com/Wahab-039/ChatApp/api/handlers/auth"
	"github.com/Wahab-039/ChatApp/api/handlers/groupmessages"
	"github.com/Wahab-039/ChatApp/api/handlers/groups"
	"github.com/Wahab-039/ChatApp/api/handlers/messages"
	"github.com/Wahab-039/ChatApp/api/handlers/users"
	"github.com/Wahab-039/ChatApp/api/middleware"
	"github.com/Wahab-039/ChatApp/api/routes"
	"github.com/Wahab-039/ChatApp/ent"
	"github.com/Wahab-039/ChatApp/internal/config"
	appmqtt "github.com/Wahab-039/ChatApp/internal/mqtt"
	groupmessagerepository "github.com/Wahab-039/ChatApp/internal/repositories/groupmessages"
	grouprepository "github.com/Wahab-039/ChatApp/internal/repositories/groups"
	messagerepository "github.com/Wahab-039/ChatApp/internal/repositories/messages"
	userrepository "github.com/Wahab-039/ChatApp/internal/repositories/users"
	authservice "github.com/Wahab-039/ChatApp/internal/services/auth"
	groupmessagesservice "github.com/Wahab-039/ChatApp/internal/services/groupmessages"
	groupsservice "github.com/Wahab-039/ChatApp/internal/services/groups"
	messagesservice "github.com/Wahab-039/ChatApp/internal/services/messages"
	userservice "github.com/Wahab-039/ChatApp/internal/services/users"
	"github.com/gin-gonic/gin"
)

// newRouter wires up all dependencies and returns a configured gin router.
func newRouter(entClient *ent.Client, cfg *config.Config, publisher *appmqtt.Publisher) *gin.Engine {
	router := gin.Default()

	userRepository := userrepository.NewEntRepository(entClient)
	messageRepository := messagerepository.NewEntRepository(entClient)
	groupRepository := grouprepository.NewEntRepository(entClient)
	groupMessageRepository := groupmessagerepository.NewEntRepository(entClient)
	tokenManager := authservice.NewTokenManager(cfg.JWTSecret, cfg.JWTAccessTTL)
	authService := authservice.NewService(userRepository, tokenManager)
	userService := userservice.NewService(userRepository)
	messageService := messagesservice.NewService(userRepository, messageRepository, publisher)
	groupService := groupsservice.NewService(groupRepository, userRepository)
	groupMessageService := groupmessagesservice.NewService(groupRepository, groupMessageRepository, publisher)
	authMiddleware := middleware.NewAuth(tokenManager)

	routes.Register(
		router,
		auth.NewHandler(authService),
		users.NewHandler(userService),
		messages.NewHandler(messageService),
		groups.NewHandler(groupService),
		groupmessages.NewHandler(groupMessageService),
		authMiddleware.RequireAuth(),
	)

	return router
}
