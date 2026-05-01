package cli

import (
	"flag"
	"fmt"
	"log"
	"os"

	"cromia/api/internal/config"
	"cromia/api/internal/db"
	"cromia/api/internal/security"
)

func RunKeys(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cromia keys <subcommand> [flags]")
		fmt.Println("Subcommands: generate, revoke")
		os.Exit(1)
	}

	cfg := config.Load()
	database, err := db.NewDB(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}

	switch args[0] {
	case "generate":
		fs := flag.NewFlagSet("keys generate", flag.ExitOnError)
		user := fs.String("user", "", "Username")
		name := fs.String("name", "Default Key", "Key name")
		fs.Parse(args[1:])

		u, err := database.GetUserByUsername(*user)
		if err != nil {
			fmt.Printf("User not found: %v\n", err)
			os.Exit(1)
		}

		keyString, keyHash, err := security.GenerateAPIKey()
		id, err := database.CreateKey(u.ID, *name, keyHash)
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("API Key generated successfully!\n")
		fmt.Printf("ID: %d\n", id)
		fmt.Printf("Key: %s\n", keyString)
		fmt.Println("SAVE THIS KEY NOW. You won't be able to see it again.")

	case "revoke":
		fs := flag.NewFlagSet("keys revoke", flag.ExitOnError)
		id := fs.Int("id", 0, "Key ID")
		fs.Parse(args[1:])

		err := database.RevokeKey(*id)
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
		} else {
			fmt.Println("Key revoked successfully.")
		}
	}
}
