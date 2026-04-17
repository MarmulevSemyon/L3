package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	// Драйвер PostgreSQL подключается через blank import, чтобы зарегистрироваться в database/sql.
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/logger"

	"eventBooker/internal/config"
	"eventBooker/internal/handler"
	"eventBooker/internal/repository"
	"eventBooker/internal/router"
	"eventBooker/internal/scheduler"
	"eventBooker/internal/service"
)

const migrationsDir = "migrations"

// App управляет жизненным циклом приложения.
type App struct {
	cfg        *config.Config
	log        logger.Logger
	db         *dbpg.DB
	httpServer *http.Server
	scheduler  *scheduler.Scheduler
}

// New создаёт экземпляр приложения, инициализирует конфигурацию, логгер, базу данных, сервисы и HTTP-сервер.
func New(cfg *config.Config) (*App, error) {
	app := &App{cfg: cfg}

	logg, err := logger.InitLogger(
		cfg.Logger.LogEngine(),
		"EventBooker",
		cfg.Gin.Mode,
		logger.WithLevel(cfg.Logger.LogLevel()),
	)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}
	app.log = logg

	if err = app.runMigrations(); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	if err = app.initDB(); err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}

	if err = app.initServices(); err != nil {
		return nil, fmt.Errorf("init services: %w", err)
	}

	return app, nil
}

func (a *App) runMigrations() error {
	db, err := sql.Open("postgres", a.cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("open db for migrations: %w", err)
	}
	defer db.Close()

	if err = goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

func (a *App) initDB() error {
	db, err := dbpg.New(
		a.cfg.Postgres.DSN(),
		nil,
		&dbpg.Options{
			MaxOpenConns:    a.cfg.Postgres.MaxOpenConns,
			MaxIdleConns:    a.cfg.Postgres.MaxIdleConns,
			ConnMaxLifetime: a.cfg.Postgres.ConnMaxLifetime,
		},
	)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}

	if err = db.Master.PingContext(context.Background()); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	a.db = db
	return nil
}

func (a *App) initServices() error {
	eventRepo := repository.NewEventRepo(a.db)
	bookingRepo := repository.NewBookingRepo(a.db)
	userRepo := repository.NewUserRepo(a.db)

	eventService := service.NewEventService(eventRepo, bookingRepo)
	userService := service.NewUserService(userRepo)
	bookingService := service.NewBookingService(bookingRepo, eventRepo, userRepo, a.log)

	a.scheduler = scheduler.New(bookingService, a.cfg.Scheduler.Interval, a.log)

	h := handler.NewHandler(eventService, bookingService, userService)
	r := router.InitRouter(a.cfg.Gin.Mode, h)

	a.httpServer = &http.Server{
		Addr:         a.cfg.Server.Addr,
		Handler:      r,
		ReadTimeout:  a.cfg.Server.ReadTimeout,
		WriteTimeout: a.cfg.Server.WriteTimeout,
		IdleTimeout:  a.cfg.Server.IdleTimeout,
	}

	return nil
}

// Run запускает HTTP-сервер, планировщик фоновых задач и ожидает завершения приложения.
func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go a.scheduler.Start(ctx)

	errCh := make(chan error, 1)
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	return a.shutdown()
}

func (a *App) shutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.Server.WriteTimeout)
	defer cancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http: %w", err)
	}

	if err := a.db.Master.Close(); err != nil {
		return fmt.Errorf("close db: %w", err)
	}

	return nil
}
