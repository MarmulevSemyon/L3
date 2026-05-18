package app

import (
	"log"
	"time"

	appconfig "warehousecontrol/internal/config"
	"warehousecontrol/internal/repository"
	"warehousecontrol/internal/service"
	httptransport "warehousecontrol/internal/transport/http"

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

	userRepo := repository.NewUserRepository(db)
	itemRepo := repository.NewItemRepository(db)
	historyRepo := repository.NewHistoryRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWT.Secret)
	itemService := service.NewItemService(itemRepo)
	historyService := service.NewHistoryService(historyRepo)

	h := httptransport.NewHandler(authService, itemService, historyService)

	router := ginext.New(cfg.HTTP.Mode)
	router.Use(ginext.Logger(), ginext.Recovery())

	router.GET("/", h.Index)
	router.GET("/health", h.Health)
	router.POST("/login", h.Login)

	api := router.Group("/api")
	api.Use(httptransport.AuthMiddleware(cfg.JWT.Secret))
	api.GET("/me", h.Me)

	items := router.Group("/items")
	items.Use(httptransport.AuthMiddleware(cfg.JWT.Secret))
	items.POST("", h.CreateItem)
	items.GET("", h.ListItems)
	items.GET("/:id/history", h.GetItemHistory)
	items.GET("/:id", h.GetItem)
	items.PUT("/:id", h.UpdateItem)
	items.DELETE("/:id", h.DeleteItem)

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
