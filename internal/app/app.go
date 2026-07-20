// Package app assembles the application and manages its lifecycle.
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Wahab-039/ChatApp/internal/config"
	"github.com/Wahab-039/ChatApp/internal/database"
	"github.com/gin-gonic/gin"
)

const databaseStartupTimeout = 5 * time.Second

// Application owns the long-lived resources required to run the API.
type Application struct {
	config   *config.Config
	database *database.Postgres
	router   *gin.Engine
}

// New creates the application's long-lived resources and configures its routes.
func New(cfg *config.Config) (*Application, error) {
	ctx, cancel := context.WithTimeout(context.Background(), databaseStartupTimeout)
	defer cancel()

	conn, err := database.NewPostgres(ctx, cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	return &Application{
		config:   cfg,
		database: conn,
		router:   newRouter(conn, cfg),
	}, nil
}

// Run starts the HTTP server.
func (a *Application) Run() error {
	return a.router.Run(":" + a.config.Port)
}

// Close releases the application's long-lived resources.
func (a *Application) Close() {
	a.database.Close()
}
