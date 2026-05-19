package db

import (
	"database/sql"
	"log"
)

func (d *sqlDB) CreateKey(userID int, name, keyHash string) (int, error) {
	result, err := d.conn.Exec(
		"INSERT INTO api_keys (user_id, name, key_hash) VALUES (?, ?, ?)",
		userID, name, keyHash,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (d *sqlDB) GetActiveKeys() ([]APIKey, error) {
	rows, err := d.conn.Query(`
		SELECT id, user_id, name, key_hash, created_at 
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
		if err := rows.Scan(&ak.ID, &ak.UserID, &ak.Name, &ak.KeyHash, &ak.CreatedAt); err != nil {
			log.Printf("[DB] Error scanning APIKey: %v", err)
			continue
		}
		keys = append(keys, ak)
	}
	return keys, nil
}

func (d *sqlDB) RevokeKey(id int) error {
	_, err := d.conn.Exec("UPDATE api_keys SET revoked_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

func (d *sqlDB) GetUserKeys(userID int) ([]APIKey, error) {
	rows, err := d.conn.Query(`
		SELECT id, user_id, name, key_hash, created_at 
		FROM api_keys 
		WHERE user_id = ? AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var ak APIKey
		if err := rows.Scan(&ak.ID, &ak.UserID, &ak.Name, &ak.KeyHash, &ak.CreatedAt); err != nil {
			log.Printf("[DB] Error scanning APIKey for user %d: %v", userID, err)
			continue
		}
		keys = append(keys, ak)
	}
	return keys, nil
}

func (d *sqlDB) GetKeyByHash(keyHash string) (*APIKey, error) {
	row := d.conn.QueryRow(`
		SELECT id, user_id, name, key_hash, created_at, revoked_at
		FROM api_keys
		WHERE key_hash = ? AND revoked_at IS NULL
	`, keyHash)

	var ak APIKey
	var revokedAt sql.NullString
	if err := row.Scan(&ak.ID, &ak.UserID, &ak.Name, &ak.KeyHash, &ak.CreatedAt, &revokedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	if revokedAt.Valid {
		ak.RevokedAt = &revokedAt.String
	}
	return &ak, nil
}
