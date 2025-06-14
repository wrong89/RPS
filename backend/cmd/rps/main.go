package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"rps/internal/config"
	"rps/internal/domain/logic/auth"
	"rps/internal/http-server/handlers/authHandler"
	"rps/internal/http-server/handlers/playerHandler"
	authMiddleware "rps/internal/http-server/middleware/auth"
	mwLogger "rps/internal/http-server/middleware/logger"
	"rps/internal/lib/logger/handlers/slogpretty"
	"rps/internal/storage/postgres"
	"syscall"
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
	loadEnv()

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	storage, err := postgres.New(context.TODO())
	if err != nil {
		log.Error("failed to initialize storage")
		panic(err)
	}

	auth := auth.NewAuthService(log, storage, storage, os.Getenv("JWT_SECRET"), time.Duration(10*time.Hour))

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

	router.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.LoginHandler(log, auth))
			r.Post("/register", authHandler.RegisterHandler(log, auth))
			// r.Post("/refresh", authHandler.RefreshTokenHandler(log, auth))
		})

		r.With(authMiddleware.AuthMiddleware(log, auth)).Get("/profile", playerHandler.ProfileHandler(log, storage))
	})

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	log.Info("starting server", slog.String("addr", cfg.HTTPServer.Address))

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	sign := <-stop

	log.Info("Gracefully Shutdown", slog.String("signal", sign.String()))

	storage.CloseDb()

	log.Info("server stopped")
	os.Exit(0)
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		panic("env not found")
	}

	requiredVars := []string{"CONFIG_PATH", "DATABASE_URL", "JWT_TOKEN_SECRET"}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Fatalf("Required environment variable %s is not set", v)
		}
	}
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
