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

// newRouter wires up all dependencies and returns a configured gin router.
// conn is the existing pgx pool (used by current repositories).
// entClient is the Ent ORM client — repositories will be migrated to use it in Phase 5.
func newRouter(conn *database.Postgres, entClient *ent.Client, cfg *config.Config, publisher *appmqtt.Publisher) *gin.Engine {
	router := gin.Default()

	// TODO Phase 5: swap these one by one to NewEntRepository(entClient)
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
