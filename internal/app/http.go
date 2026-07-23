package app

import (
	"github.com/Wahab-039/ChatApp/api/handlers"
	"github.com/Wahab-039/ChatApp/api/middleware"
	"github.com/Wahab-039/ChatApp/api/routes"
	"github.com/Wahab-039/ChatApp/internal/config"
	"github.com/Wahab-039/ChatApp/internal/database"
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

func newRouter(conn *database.Postgres, cfg *config.Config, publisher *appmqtt.Publisher) *gin.Engine {
	router := gin.Default()

	userRepository := userrepository.NewPostgresRepository(conn.Pool)
	messageRepository := messagerepository.NewPostgresRepository(conn.Pool)
	groupRepository := grouprepository.NewPostgresRepository(conn.Pool)
	groupMessageRepository := groupmessagerepository.NewPostgresRepository(conn.Pool)
	tokenManager := authservice.NewTokenManager(cfg.JWTSecret, cfg.JWTAccessTTL)
	authService := authservice.NewService(userRepository, tokenManager)
	userService := userservice.NewService(userRepository)
	messageService := messagesservice.NewService(userRepository, messageRepository, publisher)
	groupService := groupsservice.NewService(groupRepository, userRepository)
	groupMessageService := groupmessagesservice.NewService(groupRepository, groupMessageRepository, publisher)
	authMiddleware := middleware.NewAuth(tokenManager)
	loginRateLimiter := middleware.NewLoginRateLimiter(cfg.LoginRateLimit, cfg.LoginRateWindow)

	var mqttDev *handlers.MQTTDev
	if cfg.Environment == "development" {
		mqttDev = handlers.NewMQTTDev(publisher)
	}

	routes.Register(
		router,
		handlers.NewHealth(conn.Pool),
		handlers.NewAuth(authService, userService),
		handlers.NewMessages(messageService),
		handlers.NewGroups(groupService, groupMessageService),
		loginRateLimiter.LimitLogin(),
		authMiddleware.RequireAuth(),
		mqttDev,
	)

	return router
}
