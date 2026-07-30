// Package app assembles the application and manages its lifecycle.
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Wahab-039/ChatApp/ent"
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
	entClient *ent.Client // Ent ORM client — used by Ent-based repositories
	publisher *appmqtt.Publisher
	router    *gin.Engine
}

// New creates the application's long-lived resources and configures its routes.
func New(cfg *config.Config) (*Application, error) {
	ctx, cancel := context.WithTimeout(context.Background(), databaseStartupTimeout)
	defer cancel()

	// Existing pgx pool — still used by current repositories
	conn, err := database.NewPostgres(ctx, cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	// Ent client — connects to the same DB via pgx stdlib driver
	// We initialise it here alongside pgx; both can run simultaneously.
	entClient, err := database.NewEntClient(cfg.DatabaseURL())
	if err != nil {
		conn.Close() // clean up pgx pool before returning
		return nil, fmt.Errorf("connect ent: %w", err)
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
		_ = entClient.Close()
		return nil, fmt.Errorf("connect mqtt: %w", err)
	}

	return &Application{
		config:    cfg,
		database:  conn,
		entClient: entClient,
		publisher: publisher,
		router:    newRouter(conn, entClient, cfg, publisher),
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
	// Close Ent client before the pgx pool so any in-flight Ent
	// queries can finish before the underlying connections disappear.
	if a.entClient != nil {
		_ = a.entClient.Close()
	}
	a.database.Close()
}
