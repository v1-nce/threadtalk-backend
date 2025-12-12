package db

import (
	"database/sql"
	"fmt"
	"os"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Connect() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to catabase")
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("unable to ping database")
	}
	return conn,nil;
}

func RunDBMigrations(dbURL string) {
	m, err := migrate.New("file://internal/db/migrations", dbURL + "?sslmode=disable")
	if err != nil {
		log.Fatalf("Migration failed due to: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration up failed due to: %v", err)
	}
	log.Println("Migrations applied successfully")
}
