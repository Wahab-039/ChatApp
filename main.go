package main

import (
	"context"
	"log"
	"time"

	"github.com/Wahab-039/ChatApp/api/routes"
	"github.com/Wahab-039/ChatApp/internal/config"
	"github.com/Wahab-039/ChatApp/internal/database"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := database.NewPostgres(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer conn.Close()

	router := gin.Default()
	routes.Register(router)

	log.Printf("ChatApp started in %s mode on port %s", cfg.Environment, cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
