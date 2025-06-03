package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"rps/internal/config"
	"rps/internal/domain/logic/auth"
	"rps/internal/http-server/handlers/authHandler"
	mwLogger "rps/internal/http-server/middleware/logger"
	"rps/internal/lib/logger/handlers/slogpretty"
	"rps/internal/storage/postgres"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("env not found")
	}

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	storage, err := postgres.New(context.TODO())
	if err != nil {
		log.Error("failed to initialize storage")
		panic(err)
	}

	auth := auth.New(log, storage, storage, time.Hour)

	log.Info(
		"starting RPS app",
		slog.String("env", cfg.Env),
	)
	log.Debug("debug messages are enabled")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/login", authHandler.LoginHandler(log, auth))
	router.Post("/register", authHandler.RegisterHandler(log, auth))

	log.Info("starting server", slog.String("addr", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
