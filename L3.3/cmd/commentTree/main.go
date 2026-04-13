package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"commentTree/internal/config"
	"commentTree/internal/handler"
	"commentTree/internal/repository"
	"commentTree/internal/service"
	"commentTree/internal/transport"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.Postgres.DSN)
	if err != nil {
		log.Fatalf("open postgres connection: %v", err)
	}
	defer db.Close()

	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		log.Println("waiting for database...")
		time.Sleep(2 * time.Second)
	}

	repo := repository.NewPostgres(db)
	svc := service.New(repo)
	h := handler.New(svc)
	router := transport.NewHTTPHandler(h)

	addr := cfg.App.Host + ":" + cfg.App.Port

	log.Printf("server started on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("start http server: %v", err)
	}
}
