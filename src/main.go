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
    "github.com/emailimmunity/passwordimmunity/api/handlers"
    "github.com/emailimmunity/passwordimmunity/db"
    "github.com/emailimmunity/passwordimmunity/services/featureflag"
    "github.com/emailimmunity/passwordimmunity/services/licensing"
    "github.com/emailimmunity/passwordimmunity/services/payment"
    "github.com/mollie/mollie-api-go/v2/mollie"
)

type Config struct {
    ServerAddr  string
    DatabaseURL string
    MollieKey   string
}

func main() {
    // Load configuration
    cfg := loadConfig()

    // Initialize database connection
    dbConn, err := initDB(cfg)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    // Initialize repositories
    repo := db.NewRepository(dbConn)

    // Initialize Mollie client
    mollieClient := mollie.NewClient(cfg.MollieKey, nil)

    // Initialize services
    paymentService := payment.NewService(repo, mollieClient)
    licenseService := licensing.NewService(repo)
    featureFlagService := featureflag.NewService(licenseService)

    // Initialize handlers
    paymentHandler := handlers.NewPaymentHandler(paymentService, licenseService)
    licenseHandler := handlers.NewLicenseHandler(licenseService, featureFlagService)

    // Setup HTTP server
    router := api.NewRouter(paymentHandler, licenseHandler)
    srv := &http.Server{
        Addr:         cfg.ServerAddr,
        Handler:      router,
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

func loadConfig() *Config {
    return &Config{
        ServerAddr:  getEnv("SERVER_ADDR", ":8000"),
        DatabaseURL: getEnv("DATABASE_URL", "postgresql://localhost/passwordimmunity?sslmode=disable"),
        MollieKey:   getEnv("MOLLIE_API_KEY", ""),
    }
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}

func initDB(cfg *Config) (*db.DB, error) {
    return db.Connect(cfg.DatabaseURL)
}
