package main

import (
	"log"

	"eventBooker/internal/app"
	"eventBooker/internal/config"
)

func main() {
	cfg, err := config.Load("config/config.yml")
	if err != nil {
		log.Fatalf("config init: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("app init: %v", err)
	}

	if err = application.Run(); err != nil {
		log.Fatalf("app run: %v", err)
	}
}
