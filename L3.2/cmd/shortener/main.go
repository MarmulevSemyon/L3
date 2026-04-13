package main

import (
	"errors"
	"net/http"

	"Shortener/internal/repository"
	"Shortener/internal/router"
	"Shortener/internal/router/handlers"
	"Shortener/internal/service"
	"Shortener/pkg/logger"

	"github.com/wb-go/wbf/config"
	"go.uber.org/zap"
)

func main() {
	cfg := config.New()
	_ = cfg.LoadConfigFiles("./config/config.yaml")

	log, err := logger.NewLogger(cfg.GetString("log_level"))
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	repo, err := repository.NewRepository(
		cfg.GetString("master_dsn"),
		cfg.GetStringSlice("slaveDSNs"),
		log,
	)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}

	svc := service.NewService(repo, log)
	h := handlers.NewHandlerShortener(svc)
	r := router.NewRouter(cfg.GetString("log_level"), h, log)

	srv := &http.Server{
		Addr:    cfg.GetString("addr"),
		Handler: r.GetEngine(),
	}

	log.Info("starting server", zap.String("addr", srv.Addr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("failed to listen and serve", zap.Error(err))
	}
}
