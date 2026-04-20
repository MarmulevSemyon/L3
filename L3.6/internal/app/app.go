package app

import (
	"log"
	"time"

	appconfig "salestracker/internal/config"
	"salestracker/internal/repository/postgres"
	"salestracker/internal/service"
	httptransport "salestracker/internal/transport/http"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
)

// MustRun запускает приложение или завершает его с ошибкой.
func MustRun() {
	cfg := appconfig.MustLoad()
	log.Printf("http mode: %s", cfg.HTTP.Mode)
	log.Printf("db dsn: %s", cfg.DB.MasterDSN)
	db, err := dbpg.New(cfg.DB.MasterDSN, cfg.DB.SlaveDSNs, &dbpg.Options{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		log.Fatalf("ошибка подключения к базе данных: %v", err)
	}

	defer closeDB(db)

	repo := postgres.NewItemRepository(db)
	svc := service.NewItemService(repo)
	h := httptransport.NewHandler(svc)

	router := ginext.New(cfg.HTTP.Mode)
	router.Use(ginext.Logger(), ginext.Recovery())

	router.GET("/", h.Index)

	router.GET("/items/export", h.ExportCSV)
	router.GET("/items", h.ListItems)
	router.GET("/items/:id", h.GetItem)
	router.POST("/items", h.CreateItem)
	router.PUT("/items/:id", h.UpdateItem)
	router.DELETE("/items/:id", h.DeleteItem)

	router.GET("/analytics", h.GetAnalytics)

	addr := ":" + cfg.HTTP.Port
	log.Printf("server started on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("ошибка запуска HTTP-сервера: %v", err)
	}
}

func closeDB(db *dbpg.DB) {
	if db == nil {
		return
	}

	if db.Master != nil {
		_ = db.Master.Close()
	}

	for _, slave := range db.Slaves {
		if slave != nil {
			_ = slave.Close()
		}
	}
}
