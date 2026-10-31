package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/raffy-io/taskifyv3/internal/auth"
	taskifyDB "github.com/raffy-io/taskifyv3/internal/db"
	"github.com/raffy-io/taskifyv3/internal/middleware"
	"github.com/raffy-io/taskifyv3/internal/task"
)

var sessionManager *scs.SessionManager

func main() {
	ctx := context.Background()
	//--- Load ENV ---//
	_ = godotenv.Load()

	//--- Init DB ---//
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatalf("DB_URL environment variable is required")
	}

	//--- Create Connection Pool for SQLC ---//
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("Failed to create connection pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := taskifyDB.New(pool)
	

	//--- Initialize SCS Session Manager ---//
	sessionDB := stdlib.OpenDBFromPool(pool)
	sessionManager = scs.New()
	sessionManager.Store = postgresstore.New(sessionDB)
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteStrictMode

	// routes
	r := chi.NewRouter()
	r.Use(sessionManager.LoadAndSave)

	// Public routes
	r.Group(func(r chi.Router) {
		r.Mount("/auth", auth.AuthRoutes(queries,sessionManager))
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(sessionManager))
		r.Mount("/task", task.TaskRoutes(queries,sessionManager))
	})

	// server
	port := os.Getenv("PORT")
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,  // Protects against Slowloris attacks
		ReadTimeout:       15 * time.Second, // Total time to read headers + body
		WriteTimeout:      15 * time.Second, // Total time to write the response
		IdleTimeout:       60 * time.Second, // Keep-alive connection cleanup
	}

	slog.Info("Server is running", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}

}
