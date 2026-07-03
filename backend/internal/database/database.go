package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"latrode-fusion/internal/config"
)

type DB struct {
	DB *sql.DB
}

func Connect(cfg *config.Config) (*DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to ping: %w", err)
	}

	return &DB{DB: db}, nil
}

func (db *DB) Close() {
	db.DB.Close()
}
