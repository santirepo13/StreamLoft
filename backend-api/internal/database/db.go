package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	cfg.MaxConns = 25
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}
	log.Info().Msg("database connected")
	return &DB{pool: pool}, nil
}

func (db *DB) Pool() *pgxpool.Pool { return db.pool }

func (db *DB) Close() { db.pool.Close() }

func (db *DB) RunMigrations(ctx context.Context) error {
	migs := []string{
		`CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, numeric_id VARCHAR(20) UNIQUE, name VARCHAR(100), stream_key VARCHAR(32) UNIQUE, bitrate INTEGER, created_at TIMESTAMP DEFAULT NOW(), updated_at TIMESTAMP DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS destinations (id SERIAL PRIMARY KEY, name VARCHAR(100) UNIQUE, rtmp_url VARCHAR(500), created_at TIMESTAMP DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS user_destinations (id SERIAL PRIMARY KEY, user_id INT, destination_id INT, stream_key TEXT, enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)), created_at TIMESTAMP DEFAULT NOW(), updated_at TIMESTAMP DEFAULT NOW(), UNIQUE(user_id, destination_id))`,
		`CREATE TABLE IF NOT EXISTS broadcast_sessions (id SERIAL PRIMARY KEY, user_id INT, user_destination_id INT, date DATE DEFAULT CURRENT_DATE, duration_minutes INT DEFAULT 0, started_at TIMESTAMP DEFAULT NOW(), ended_at TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS user_machines (id SERIAL PRIMARY KEY, machine_id VARCHAR(255), user_id INT, last_used_at TIMESTAMP DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS user_sessions (id SERIAL PRIMARY KEY, user_id INT, machine_id VARCHAR(255), access_token TEXT, refresh_token TEXT, token_expires_at TIMESTAMP, created_at TIMESTAMP DEFAULT NOW(), updated_at TIMESTAMP DEFAULT NOW())`,
	}
	for _, m := range migs {
		if _, err := db.pool.Exec(ctx, m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Info().Msg("migrations done")
	return nil
}