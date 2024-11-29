package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emailimmunity/passwordimmunity/api/handlers"
	"github.com/emailimmunity/passwordimmunity/api/router"
	"github.com/emailimmunity/passwordimmunity/services/featureflag"
	"github.com/emailimmunity/passwordimmunity/services/payment"
	"github.com/emailimmunity/passwordimmunity/db"
	"github.com/emailimmunity/passwordimmunity/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize services
	featureManager := featureflag.NewFeatureManager(database)

	// Initialize payment service with configuration
	paymentService, err := payment.NewService(payment.Config{
		MollieAPIKey:    cfg.MollieAPIKey,
		WebhookBaseURL:  cfg.WebhookBaseURL,
		Database:        database,
	})
	if err != nil {
		log.Fatalf("Failed to initialize payment service: %v", err)
	}

	// Initialize enterprise feature handler
	enterpriseFeatureHandler := handlers.NewEnterpriseFeatureHandler(featureManager, paymentService)

	// Create router with all services
	r := router.NewRouter(featureManager, paymentService, enterpriseFeatureHandler)

	// Configure server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
