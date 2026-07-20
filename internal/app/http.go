package app

import (
	"github.com/Wahab-039/ChatApp/api/handlers"
	"github.com/Wahab-039/ChatApp/api/middleware"
	"github.com/Wahab-039/ChatApp/api/routes"
	"github.com/Wahab-039/ChatApp/internal/config"
	"github.com/Wahab-039/ChatApp/internal/database"
	userrepository "github.com/Wahab-039/ChatApp/internal/repositories/users"
	authservice "github.com/Wahab-039/ChatApp/internal/services/auth"
	userservice "github.com/Wahab-039/ChatApp/internal/services/users"
	"github.com/gin-gonic/gin"
)

func newRouter(conn *database.Postgres, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	userRepository := userrepository.NewPostgresRepository(conn.Pool)
	tokenManager := authservice.NewTokenManager(cfg.JWTSecret, cfg.JWTAccessTTL)
	authService := authservice.NewService(userRepository, tokenManager)
	userService := userservice.NewService(userRepository)
	authMiddleware := middleware.NewAuth(tokenManager)
	loginRateLimiter := middleware.NewLoginRateLimiter(cfg.LoginRateLimit, cfg.LoginRateWindow)

	routes.Register(
		router,
		handlers.NewHealth(conn.Pool),
		handlers.NewAuth(authService, userService),
		loginRateLimiter.LimitLogin(),
		authMiddleware.RequireAuth(),
	)

	return router
}
