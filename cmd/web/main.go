package main

import (
	"context"
	"database/sql"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/raffy-io/taskifyv3"
	"github.com/raffy-io/taskifyv3/internal/auth"
	taskifyDB "github.com/raffy-io/taskifyv3/internal/db"
	"github.com/raffy-io/taskifyv3/internal/home"
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

	//--- Database Connection for Migrations & Sessions ---//
	// Goose and SCS both use standard database/sql connections (*sql.DB)
	sessionDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}
	defer sessionDB.Close()

	// Run auto-migrations on startup
	if err := taskifyv3.RunMigrations(sessionDB); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("Migrations applied successfully!")

	//--- Create Connection Pool for SQLC ---//
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("Failed to create connection pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := taskifyDB.New(pool)

	//--- Initialize SCS Session Manager ---//
	sessionDB = stdlib.OpenDBFromPool(pool)
	sessionManager = scs.New()
	sessionManager.Store = postgresstore.New(sessionDB)
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteStrictMode

	// Automatically loads from the CLOUDINARY_URL environment variable
	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}

	// routes
	r := chi.NewRouter()

	// --- 1. Global Middlewares ---
    // Initialize rate limiter: e.g., 20 requests per interval check, with burst capacity
    limiter := middleware.NewRateLimiter(20, 50, time.Minute)
    
    r.Use(limiter.Middleware)           // Protect against spam/DDoS
    r.Use(middleware.UserAgentValidator) // Block empty/blank scrapers

	// -----static assets CSS JS IMG----- //
	staticFS, err := fs.Sub(taskifyv3.EmbeddedAssets, "static")
	if err != nil {
		slog.Error("Cant find static assets", "error", err)
		os.Exit(1)
	}
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Public routes
	r.Group(func(r chi.Router) {
		r.Use(sessionManager.LoadAndSave)
		r.Use(middleware.NoSurf)
		r.Mount("/auth", auth.AuthRoutes(queries, sessionManager))
		r.Mount("/", home.HomeRoutes(queries, sessionManager,cld))
	})

	// Protected user routes
	r.Group(func(r chi.Router) {
		r.Use(sessionManager.LoadAndSave)
		r.Use(middleware.RequireAuth(sessionManager))
		r.Use(middleware.NoSurf)
		r.Mount("/task", task.TaskRoutes(queries, sessionManager))
	})

	// Protected admin routes
	r.Group(func(r chi.Router) {
		r.Use(sessionManager.LoadAndSave)
		r.Use(middleware.RequireAuth(sessionManager))
		r.Use(middleware.RequireAdmin(sessionManager))
		r.Use(middleware.NoSurf)

		r.Mount("/admin",home.AdminRoutes(queries,sessionManager,cld))
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
