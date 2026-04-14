package main

import (
	"context"
	"database/sql"
	"imageProcessor/internal/broker"
	"imageProcessor/internal/config"
	"imageProcessor/internal/handler"
	"imageProcessor/internal/repository"
	"imageProcessor/internal/service"
	"imageProcessor/internal/storage"
	"imageProcessor/internal/transport"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/wb-go/wbf/logger"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	appLogger, err := logger.InitLogger(
		logger.ZerologEngine,
		"imageProcessor",
		"dev",
		logger.WithLevel(logger.InfoLevel),
	)
	if err != nil {
		log.Fatalf("logger init error: %v", err)
	}

	db, err := sql.Open("postgres", cfg.Postgres.DSN)
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping error: %v", err)
	}

	fileStore := storage.NewFileStore(
		cfg.Storage.OriginalDir,
		cfg.Storage.ProcessedDir,
		cfg.Storage.ThumbsDir,
	)

	if err := fileStore.EnsureDirs(); err != nil {
		log.Fatalf("storage init error: %v", err)
	}

	repo := repository.NewPostgresRepository(db)

	producer := broker.NewKafkaProducer(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topic,
		appLogger,
	)
	defer producer.Close()

	svc := service.New(repo, fileStore, producer)
	h := handler.New(svc)
	router := transport.NewRouter(h)

	consumer, err := broker.NewConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topic,
		cfg.Kafka.GroupID,
		cfg.Kafka.DLQTopic,
		appLogger,
		svc,
	)
	if err != nil {
		log.Fatalf("consumer init error: %v", err)
	}
	defer consumer.Close()

	server := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		appLogger.LogAttrs(context.Background(), logger.InfoLevel, "http server started",
			logger.String("port", cfg.App.Port),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()

	appLogger.LogAttrs(context.Background(), logger.InfoLevel, "shutdown signal received")

	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
