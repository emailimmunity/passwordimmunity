package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emailimmunity/passwordimmunity/api"
	"github.com/emailimmunity/passwordimmunity/db"
)

func main() {
	// Load configuration
	cfg := loadConfig()

	// Initialize database connection
	if err := initDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup HTTP server
	srv := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      api.SetupRoutes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", cfg.ServerAddr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

type Config struct {
	ServerAddr  string
	DatabaseURL string
	// Add other configuration fields as needed
}

func loadConfig() *Config {
	return &Config{
		ServerAddr:  getEnv("SERVER_ADDR", ":8000"),
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://localhost/passwordimmunity?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func initDB(cfg *Config) error {
	// Database initialization will be implemented in a separate PR
	return nil
}
