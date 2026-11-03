package taskifyv3

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/pressly/goose/v3"
)

//go:embed db/migration/*.sql
var migrationFiles embed.FS

func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrationFiles)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, "db/migration"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// --- Seed Admin from Environment Variables ---
	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "Admin" // fallback default
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@taskify.com" // fallback default
	}

	adminPasswordHash := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD_HASH"))
	slog.Info("DEBUG: Checking admin password hash", "length", len(adminPasswordHash), "hash", adminPasswordHash)
	// If not set in env, you might want to skip or log a warning
	if adminPasswordHash != "" {
		seedQuery := `
    INSERT INTO users (username,email, password_hash, role) 
    VALUES ($1, $2,$3, 'admin')
    ON CONFLICT (email) 
    DO UPDATE SET 
    username = EXCLUDED.username,
    password_hash = EXCLUDED.password_hash, 
    role = EXCLUDED.role;
`
		_, err := db.Exec(seedQuery, adminUsername, adminEmail, adminPasswordHash)
		if err != nil {
			return fmt.Errorf("failed to seed admin user: %w", err)
		}
	}

	return nil
}
