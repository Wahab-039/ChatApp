// Package app assembles the application and manages its lifecycle.
package app

import (
	"database/sql"
	"fmt"

	"github.com/Wahab-039/ChatApp/ent"
	"github.com/Wahab-039/ChatApp/internal/config"
	"github.com/Wahab-039/ChatApp/internal/database"
	appmqtt "github.com/Wahab-039/ChatApp/internal/mqtt"
	"github.com/gin-gonic/gin"
)

// Application owns the long-lived resources required to run the API.
type Application struct {
	config    *config.Config
	entClient *ent.Client
	sqlDB     *sql.DB
	publisher *appmqtt.Publisher
	router    *gin.Engine
}

// New creates the application's long-lived resources and configures its routes.
func New(cfg *config.Config) (*Application, error) {
	entClient, sqlDB, err := database.NewEntClient(cfg.DatabaseURL())
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
		_ = entClient.Close()
		return nil, fmt.Errorf("connect mqtt: %w", err)
	}

	return &Application{
		config:    cfg,
		entClient: entClient,
		sqlDB:     sqlDB,
		publisher: publisher,
		router:    newRouter(entClient, sqlDB, cfg, publisher),
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
	if a.entClient != nil {
		// Closing entClient also closes the underlying *sql.DB.
		_ = a.entClient.Close()
	}
}
