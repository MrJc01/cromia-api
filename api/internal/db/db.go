package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type sqlDB struct {
	conn   *sql.DB
	driver string
}

func NewDB(driver, dsn string) (DB, error) {
	if driver == "sqlite3" {
		params := []string{
			"_busy_timeout=5000",
			"_journal_mode=WAL",
			"_sync=NORMAL",
		}
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		dsn += separator + strings.Join(params, "&")
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	if driver == "sqlite3" {
		db.SetMaxOpenConns(50)
		db.SetMaxIdleConns(50)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	instance := &sqlDB{conn: db, driver: driver}
	if err := instance.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	return instance, nil
}

func (d *sqlDB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			balance DECIMAL(10,4) NOT NULL DEFAULT 0.0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			key_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			revoked_at DATETIME NULL,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS provider_models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			provider_name TEXT NOT NULL,
			model_name TEXT NOT NULL UNIQUE,
			cost_multiplier DECIMAL(10,4) NOT NULL DEFAULT 1.0,
			prompt_cost REAL NOT NULL DEFAULT 0.0,
			completion_cost REAL NOT NULL DEFAULT 0.0,
			is_active BOOLEAN NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS usage_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			api_key_id INTEGER NOT NULL,
			model_used TEXT NOT NULL,
			prompt_tokens INTEGER NOT NULL,
			completion_tokens INTEGER NOT NULL,
			cost_deducted DECIMAL(10,4) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id),
			FOREIGN KEY(api_key_id) REFERENCES api_keys(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(revoked_at) WHERE revoked_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_usage_logs_user ON usage_logs(user_id)`,
	}

	for _, q := range queries {
		if _, err := d.conn.Exec(q); err != nil {
			return fmt.Errorf("query failed: %s, err: %w", q, err)
		}
	}

	// Migrações adicionais com fallback
	alterQueries := []string{
		`ALTER TABLE provider_models ADD COLUMN prompt_cost REAL NOT NULL DEFAULT 0.0`,
		`ALTER TABLE provider_models ADD COLUMN completion_cost REAL NOT NULL DEFAULT 0.0`,
		`ALTER TABLE users ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0`,
	}
	for _, q := range alterQueries {
		d.conn.Exec(q) // Ignora erro se a coluna já existir
	}

	return nil
}
