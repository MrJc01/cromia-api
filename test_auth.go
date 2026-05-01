package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"cromia/api/internal/security"
	"cromia/api/internal/db"
	"cromia/api/internal/config"
	"log"
)

func main() {
	key, hash, err := security.GenerateAPIKey()
	fmt.Printf("Key: '%s'\nHash: '%s'\nErr: %v\n", key, hash, err)

	ok, err := security.CompareAPIKey(key, hash)
	fmt.Printf("Compare OK: %v, Err: %v\n", ok, err)

	cfg := config.Load()
	database, _ := db.NewDB("sqlite3", "test_data.db")
	keys, _ := database.GetActiveKeys()
	for _, k := range keys {
		fmt.Printf("DB Key Hash: '%s'\n", k.KeyHash)
	}
}
