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

func RunUsers(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cromia users <subcommand> [flags]")
		fmt.Println("Subcommands: create, add-credits, list")
		os.Exit(1)
	}

	cfg := config.Load()
	database, err := db.NewDB(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	switch args[0] {
	case "create":
		fs := flag.NewFlagSet("users create", flag.ExitOnError)
		username := fs.String("username", "", "Username")
		password := fs.String("password", "", "Password")
		balance := fs.Float64("balance", 0.0, "Initial balance")
		fs.Parse(args[1:])

		if *username == "" || *password == "" {
			fmt.Println("Username and password are required")
			os.Exit(1)
		}

		hash, _ := security.HashPassword(*password)
		id, err := database.CreateUser(*username, hash, *balance)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("User created! ID: %d\n", id)

	case "add-credits":
		fs := flag.NewFlagSet("users add-credits", flag.ExitOnError)
		user := fs.String("user", "", "Username")
		amount := fs.Float64("amount", 0.0, "Credits to add")
		fs.Parse(args[1:])

		u, err := database.GetUserByUsername(*user)
		if err != nil {
			fmt.Printf("User not found: %v\n", err)
			os.Exit(1)
		}

		err = database.AddBalance(u.ID, *amount)
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added %.2f credits to %s\n", *amount, *user)

	case "remove-credits":
		fs := flag.NewFlagSet("users remove-credits", flag.ExitOnError)
		user := fs.String("user", "", "Username")
		amount := fs.Float64("amount", 0.0, "Credits to remove")
		fs.Parse(args[1:])

		u, err := database.GetUserByUsername(*user)
		if err != nil {
			fmt.Printf("User not found: %v\n", err)
			os.Exit(1)
		}

		err = database.DeductBalance(u.ID, *amount)
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Removed %.2f credits from %s\n", *amount, *user)

	case "list":
		users, _ := database.ListUsers()
		for _, u := range users {
			fmt.Printf("ID: %d | User: %s | Balance: %.2f\n", u.ID, u.Username, u.Balance)
		}
	default:
		fmt.Printf("Unknown user subcommand: %s\n", args[0])
	}
}
