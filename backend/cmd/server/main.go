package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	auditHandler "github.com/controlcrud/backend/internal/api/handlers/audit"
	connHandler "github.com/controlcrud/backend/internal/api/handlers/connection"
	ctrlHandler "github.com/controlcrud/backend/internal/api/handlers/controls"
	pushHandler "github.com/controlcrud/backend/internal/api/handlers/push"
	stmtHandler "github.com/controlcrud/backend/internal/api/handlers/statements"
	syncHandler "github.com/controlcrud/backend/internal/api/handlers/sync"
	"github.com/controlcrud/backend/internal/config"
	"github.com/controlcrud/backend/internal/domain/audit"
	"github.com/controlcrud/backend/internal/domain/connection"
	"github.com/controlcrud/backend/internal/domain/controls"
	"github.com/controlcrud/backend/internal/domain/pull"
	"github.com/controlcrud/backend/internal/domain/push"
	"github.com/controlcrud/backend/internal/domain/statement"
	"github.com/controlcrud/backend/internal/domain/system"
	"github.com/controlcrud/backend/internal/infrastructure/crypto"
	"github.com/controlcrud/backend/internal/infrastructure/database"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection established")

	// Initialize crypto service
	cryptoService, err := crypto.NewAESCryptoService(cfg.Encryption.Key)
	if err != nil {
		log.Fatalf("Failed to initialize crypto service: %v", err)
	}

	// Initialize logger
	logger := slog.Default()

	// Initialize repositories
	connRepo := database.NewConnectionRepository(db)
	systemRepo := database.NewSystemRepository(db)
	controlRepo := database.NewControlRepository(db)
	stmtRepo := database.NewStatementRepository(db)
	pullRepo := database.NewPullRepository(db)
	auditRepo := database.NewAuditRepository(db)

	// Initialize services
	connService := connection.NewService(connRepo, cryptoService)
	controlsService := controls.NewService(connService)
	systemService := system.NewService(systemRepo, connService, logger)
	stmtService := statement.NewService(stmtRepo, logger)
	pullService := pull.NewService(pullRepo, systemRepo, controlRepo, stmtRepo, connService, logger)
	pushService := push.NewService(stmtRepo, connService, logger)
	auditService := audit.NewService(auditRepo, logger)

	// Initialize handlers
	connectionHandler := connHandler.NewHandler(connService)
	controlsHandler := ctrlHandler.NewHandler(controlsService)
	statementsHandler := stmtHandler.NewHandler(stmtService, logger)
	syncAPIHandler := syncHandler.NewHandler(systemService, pullService, logger)
	pushAPIHandler := pushHandler.NewHandler(pushService, logger)
	auditAPIHandler := auditHandler.NewHandler(auditService, logger)

	// Create HTTP server mux
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", healthHandler(db))

	// Register connection routes
	connectionHandler.RegisterRoutes(mux)

	// Register controls routes
	controlsHandler.RegisterRoutes(mux)

	// Register statements routes
	statementsHandler.RegisterRoutes(mux)

	// Register sync routes
	syncAPIHandler.RegisterRoutes(mux)

	// Register push routes
	pushAPIHandler.RegisterRoutes(mux)

	// Register audit routes
	auditAPIHandler.RegisterRoutes(mux)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      corsMiddleware(mux),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown complete")
}

// healthHandler returns a health check handler that includes database status.
func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := "healthy"
		dbStatus := "connected"

		if err := db.PingContext(ctx); err != nil {
			status = "degraded"
			dbStatus = "disconnected"
		}

		w.Header().Set("Content-Type", "application/json")
		if status != "healthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(map[string]string{
			"status":   status,
			"database": dbStatus,
		})
	}
}

// corsMiddleware adds CORS headers for development.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from frontend dev server
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
