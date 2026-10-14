package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// Config holds PostgreSQL database connection parameters.
type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DefaultConfig returns default connection settings.
func DefaultConfig() Config {
	return Config{
		Host:            "localhost",
		Port:            "5432",
		User:            "postgres",
		Password:        "postgres",
		DBName:          "encertia",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// Connect initializes and validates a PostgreSQL connection pool.
func Connect(cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error obrint connexió postgres: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error fent ping a postgres: %w", err)
	}

	return db, nil
}

// RunMigrations applies embedded .up.sql migration scripts in order.
func RunMigrations(db *sql.DB) error {
	// 1. Ensure schema_migrations table exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("error creant taula schema_migrations: %w", err)
	}

	migrationFiles := []string{
		"migrations/000001_create_auth_tables.up.sql",
		"migrations/000002_add_admin_role_and_user_features.up.sql",
		"migrations/000003_create_quiz_tables.up.sql",
		"migrations/000004_seed_initial_users.up.sql",
		"migrations/000005_create_match_tables.up.sql",
		"migrations/000006_create_evaluation_tables.up.sql",
		"migrations/000007_create_revoked_access_tokens.up.sql",
		"migrations/000008_create_course_tables.up.sql",
		"migrations/000009_create_material_tables.up.sql",
		"migrations/000010_add_user_language.up.sql",
		"migrations/000011_create_metrics_tables.up.sql",
		"migrations/000012_create_group_tables.up.sql",
		"migrations/000013_add_group_id_to_matches.up.sql",
		"migrations/000014_allow_anonymous_match_players.up.sql",
		"migrations/000015_allow_anonymous_evaluations.up.sql",
		"migrations/000016_create_authorized_devices.up.sql",
	}

	for _, file := range migrationFiles {
		var applied int
		_ = db.QueryRow("SELECT 1 FROM schema_migrations WHERE version = $1", file).Scan(&applied)
		if applied == 1 {
			continue
		}

		content, err := MigrationsFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error llegint fitxer de migració %s: %w", file, err)
		}

		log.Printf("[DB] Executant migració: %s", file)
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("error executant migració %s: %w", file, err)
		}

		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT (version) DO NOTHING", file); err != nil {
			return fmt.Errorf("error registrant migració %s a schema_migrations: %w", file, err)
		}
	}

	log.Println("[DB] Migracions completades amb èxit.")
	return nil
}
