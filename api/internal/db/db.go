package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type APIKey struct {
	ID              int
	Name            string
	KeyHash         string
	CreatedAt       string
	RevokedAt       string
	RateLimit       int    // requests per window
	RateLimitWindow string // currently fixed to 1m, but field added for future
}

type DB interface {
	GetActiveKeys() ([]APIKey, error)
	GetAPIKeyByHash(key string) (*APIKey, error)
	SaveKey(name string, hash string) error
	CreateRequest(apiKeyID int, model string, prompt string) (int, error)
	ListModels() ([]string, error)
	RevokeKey(id int) error
}

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
		db.SetMaxIdleConns(50) // Mantém conexões prontas para performance
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


// migrate cria as tabelas necessárias caso não existam.
func (d *sqlDB) migrate() error {
	_, err := d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS api_keys (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT        NOT NULL,
			key_hash   TEXT        NOT NULL,
			created_at DATETIME    DEFAULT CURRENT_TIMESTAMP,
			revoked_at DATETIME    NULL,
			rate_limit INTEGER     NOT NULL DEFAULT 60,
			rate_limit_window TEXT NOT NULL DEFAULT '1m'
		)
	`)
	if err != nil {
		return err
	}

	// Força adição de colunas em banco legado (se falhar é porque já existe, então ignoramos o erro)
	d.conn.Exec(`ALTER TABLE api_keys ADD COLUMN rate_limit INTEGER NOT NULL DEFAULT 60`)
	d.conn.Exec(`ALTER TABLE api_keys ADD COLUMN rate_limit_window TEXT NOT NULL DEFAULT '1m'`)

	_, err = d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS requests (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			api_key_id INTEGER     NOT NULL,
			model      TEXT        NOT NULL,
			prompt     TEXT        NOT NULL,
			created_at DATETIME    DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}
	_, err = d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS models (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		)
	`)
	if err != nil {
		return err
	}

	// Garante que o modelo padrão existe
	d.conn.Exec(`INSERT OR IGNORE INTO models (name) VALUES ('crom-1')`)

	// Índices para performance
	d.conn.Exec(`CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(revoked_at) WHERE revoked_at IS NULL`)
	d.conn.Exec(`CREATE INDEX IF NOT EXISTS idx_requests_api_key ON requests(api_key_id)`)
	d.conn.Exec(`CREATE INDEX IF NOT EXISTS idx_requests_created ON requests(created_at)`)

	return nil
}

// GetAPIKeyByHash não é usado diretamente; use GetActiveKeys + security.CompareAPIKey.
func (d *sqlDB) GetAPIKeyByHash(key string) (*APIKey, error) {
	_ = key
	return nil, sql.ErrNoRows
}

// GetActiveKeys retorna todas as chaves ativas para comparação Argon2.
func (d *sqlDB) GetActiveKeys() ([]APIKey, error) {
	rows, err := d.conn.Query(`
		SELECT id, name, key_hash, created_at, COALESCE(revoked_at, ''), rate_limit, rate_limit_window
		FROM api_keys
		WHERE revoked_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var ak APIKey
		if err := rows.Scan(&ak.ID, &ak.Name, &ak.KeyHash, &ak.CreatedAt, &ak.RevokedAt, &ak.RateLimit, &ak.RateLimitWindow); err != nil {
			log.Printf("[DB] Error scanning APIKey: %v", err)
			continue
		}
		keys = append(keys, ak)
	}
	return keys, nil
}

func (d *sqlDB) SaveKey(name string, hash string) error {
	_, err := d.conn.Exec("INSERT INTO api_keys (name, key_hash) VALUES (?, ?)", name, hash)
	return err
}

func (d *sqlDB) CreateRequest(apiKeyID int, model string, prompt string) (int, error) {
	result, err := d.conn.Exec(
		"INSERT INTO requests (api_key_id, model, prompt) VALUES (?, ?, ?)",
		apiKeyID, model, prompt,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (d *sqlDB) ListModels() ([]string, error) {
	rows, err := d.conn.Query("SELECT name FROM models")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var models []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		models = append(models, name)
	}
	return models, nil
}

func (d *sqlDB) RevokeKey(id int) error {
	_, err := d.conn.Exec("UPDATE api_keys SET revoked_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}
