package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// User representa um usuário legado
type User struct {
	ID           int
	Email        string
	PasswordHash string
	Role         string
	Credits      int
}

// APIKey representa uma chave legada
type APIKey struct {
	ID        int
	UserID    int
	Key       string
	Name      string
	IsActive  bool
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Uso: go run migrate_users.go <caminho_do_banco_legado.sqlite>")
	}

	dbPath := os.Args[1]
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Erro ao abrir banco: %v", err)
	}
	defer db.Close()

	// 1. Extrair Usuários Legados
	rows, err := db.Query("SELECT id, email, password_hash, role, credits FROM users")
	if err != nil {
		log.Fatalf("Erro ao consultar users: %v", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Credits); err != nil {
			log.Printf("Erro ao scan user: %v", err)
			continue
		}
		users = append(users, u)
	}

	// 2. Extrair Chaves Legadas
	kRows, err := db.Query("SELECT id, user_id, api_key, name, is_active FROM api_keys")
	if err != nil {
		log.Printf("Erro ao consultar keys: %v", err)
	} else {
		defer kRows.Close()
	}

	var keys []APIKey
	if kRows != nil {
		for kRows.Next() {
			var k APIKey
			if err := kRows.Scan(&k.ID, &k.UserID, &k.Key, &k.Name, &k.IsActive); err != nil {
				continue
			}
			keys = append(keys, k)
		}
	}

	// 3. Gerar Dump SQL para o Cloud Novo (Postgres)
	outFile := "migration_dump_to_cloud.sql"
	f, err := os.Create(outFile)
	if err != nil {
		log.Fatalf("Erro ao criar arquivo: %v", err)
	}
	defer f.Close()

	f.WriteString("-- Arquivo de Migracao Gerado Automaticamente\n")
	f.WriteString("-- CromIA API -> CromIA Cloud\n\n")

	for _, u := range users {
		// O Cloud novo já deve ter a tabela "users" similar. Ajuste as colunas conforme necessário.
		stmt := fmt.Sprintf("INSERT INTO users (email, password_hash, role, plan_id, credits) VALUES ('%s', '%s', '%s', NULL, %d) ON CONFLICT (email) DO NOTHING;\n",
			u.Email, u.PasswordHash, u.Role, u.Credits)
		f.WriteString(stmt)
	}

	for _, k := range keys {
		// As chaves antigas sao inseridas na tabela de keys (precisa de subquery para pegar o novo ID do usuario)
		stmt := fmt.Sprintf("INSERT INTO api_keys (user_id, key_hash, name) SELECT id, '%s', '%s' FROM users WHERE id = %d;\n",
			k.Key, k.Name, k.UserID)
		f.WriteString(stmt)
	}

	fmt.Printf("Migração gerada com sucesso: %s\n", outFile)
	fmt.Printf("Total: %d usuários, %d chaves.\n", len(users), len(keys))
}
