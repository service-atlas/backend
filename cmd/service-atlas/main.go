package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"service-atlas/api/routes"
	"service-atlas/internal/config"
	"service-atlas/internal/secrets"
	"service-atlas/neo4jrepositories"
	"strings"
	"syscall"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	ctx := context.Background()
	logger := getLogger()
	slog.SetDefault(logger)
	secretsProvider, err := secrets.NewProvider(ctx)
	if err != nil {
		slog.Error("Error creating secrets provider: ", slog.Any("error", err))
		os.Exit(1)
	}
	dbInfo, err := secretsProvider.GetDatabaseInfo(ctx)
	if err != nil {
		slog.Error("Error getting database info: ", slog.Any("error", err))
		os.Exit(1)
	}
	driver, err := neo4j.NewDriverWithContext(
		dbInfo.URL,
		neo4j.BasicAuth(dbInfo.Username, dbInfo.Password, ""))
	if err != nil {
		slog.Error("Error creating driver: ", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		closeErr := driver.Close(ctx)
		if closeErr != nil {
			slog.Error("error closing driver: ", slog.Any("error", closeErr))
		}
	}()
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		panic(err)
	}
	//setup indexes for search
	err = neo4jrepositories.Startup(ctx, driver)
	if err != nil {
		panic(err)
	}

	mux := routes.SetupRouter(driver)

	server := &http.Server{
		Handler: mux,
		Addr:    config.GetConfigValue("address"),
	}

	slog.Info("Starting Web Server")
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen error", slog.Any("error", err))
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown:", slog.Any("error", err))
	}
}

func getLogger() *slog.Logger {
	lvlEnv, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		lvlEnv = "info"
	}
	switch strings.ToLower(lvlEnv) {
	case "debug":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case "error":
		return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError,
		}))
	case "warning":
		return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		}))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))

}
