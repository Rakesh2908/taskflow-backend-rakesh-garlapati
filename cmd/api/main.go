package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Rakesh2908/taskflow/api/handler"
	"github.com/Rakesh2908/taskflow/config"
	"github.com/Rakesh2908/taskflow/middleware"
	"github.com/Rakesh2908/taskflow/repository"
	"github.com/Rakesh2908/taskflow/routes"
	"github.com/Rakesh2908/taskflow/service"
	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		logger.Error("db pool init failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("db ping failed", "err", err)
		os.Exit(1)
	}

	if err := runMigrations(cfg.DBUrl, logger); err != nil {
		logger.Error("migrations failed", "err", err)
		os.Exit(1)
	}

	usersRepo := repository.NewUserRepository(pool)
	projectsRepo := repository.NewProjectRepository(pool)
	tasksRepo := repository.NewTaskRepository(pool)

	authSvc := service.NewAuthService(cfg, usersRepo)
	projectSvc := service.NewProjectService(projectsRepo, tasksRepo)
	taskSvc := service.NewTaskService(tasksRepo, projectsRepo)

	v := validator.New()
	authH := handler.NewAuthHandler(authSvc, v)
	projectH := handler.NewProjectHandler(projectSvc, v)
	taskH := handler.NewTaskHandler(taskSvc, projectSvc, v)

	router := routes.NewRouter(&cfg, authH, projectH, taskH)
	router = middleware.ForceJSONContentType()(router)
	router = middleware.Logger(logger)(router)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-sigCh:
		logger.Info("signal received, shutting down", "signal", sig.String())
	case err := <-errCh:
		logger.Error("server error", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}

func runMigrations(dbURL string, logger *slog.Logger) error {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}

	drv, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", drv)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("migrations up-to-date")
			return nil
		}
		return err
	}

	logger.Info("migrations applied")
	return nil
}

