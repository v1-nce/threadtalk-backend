package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}
	// Validate that sslmode is set for production
	// In production, we should enforce SSL/TLS
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	conn.SetMaxOpenConns(5)                  // Maximum open connections (Lambda-optimized)
	conn.SetMaxIdleConns(2)                  // Maximum idle connections
	conn.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime
	conn.SetConnMaxIdleTime(1 * time.Minute) // Maximum idle time before closing
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}
	return conn, nil
}

func RunDBMigrations(dbURL string) {
	m, err := migrate.New("file://internal/db/migrations", dbURL)
	if err != nil {
		log.Fatalf("Migration failed due to: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration up failed due to: %v", err)
	}
	log.Println("Migrations applied successfully")
}
