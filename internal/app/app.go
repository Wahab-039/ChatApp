// Package app assembles the application and manages its lifecycle.
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Wahab-039/ChatApp/internal/config"
	"github.com/Wahab-039/ChatApp/internal/database"
	appmqtt "github.com/Wahab-039/ChatApp/internal/mqtt"
	"github.com/gin-gonic/gin"
)

const databaseStartupTimeout = 5 * time.Second

// Application owns the long-lived resources required to run the API.
type Application struct {
	config    *config.Config
	database  *database.Postgres
	publisher *appmqtt.Publisher
	router    *gin.Engine
}

// New creates the application's long-lived resources and configures its routes.
func New(cfg *config.Config) (*Application, error) {
	ctx, cancel := context.WithTimeout(context.Background(), databaseStartupTimeout)
	defer cancel()

	conn, err := database.NewPostgres(ctx, cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	publisher, err := appmqtt.Connect(appmqtt.Config{
		BrokerURL:      cfg.MQTTBrokerURL,
		Username:       cfg.MQTTServiceUsername,
		Password:       cfg.MQTTServicePassword,
		ClientID:       cfg.MQTTClientID,
		ConnectTimeout: cfg.MQTTConnectTimeout,
	})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("connect mqtt: %w", err)
	}

	return &Application{
		config:    cfg,
		database:  conn,
		publisher: publisher,
		router:    newRouter(conn, cfg, publisher),
	}, nil
}

// Run starts the HTTP server.
func (a *Application) Run() error {
	return a.router.Run(":" + a.config.Port)
}

// Close releases the application's long-lived resources.
func (a *Application) Close() {
	if a.publisher != nil {
		a.publisher.Close()
	}
	a.database.Close()
}
