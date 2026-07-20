package main

import (
	"log"

	"github.com/Wahab-039/ChatApp/internal/app"
	"github.com/Wahab-039/ChatApp/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("create application: %v", err)
	}
	defer application.Close()

	log.Printf("ChatApp started in %s mode on port %s", cfg.Environment, cfg.Port)
	if err := application.Run(); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
