package cli

import (
	"flag"
	"fmt"
	"log"
	"os"

	"cromia/api/internal/config"
	"cromia/api/internal/db"
	"cromia/api/internal/workers"
)

func RunModels(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cromia models <subcommand> [flags]")
		fmt.Println("Subcommands: enable, disable, list, sync-pricing")
		os.Exit(1)
	}

	cfg := config.Load()
	database, err := db.NewDB(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}

	switch args[0] {
	case "enable":
		fs := flag.NewFlagSet("models enable", flag.ExitOnError)
		provider := fs.String("provider", "", "Provider (deepseek, openrouter)")
		model := fs.String("model", "", "Model name (e.g. deepseek-chat)")
		multiplier := fs.Float64("multiplier", 1.0, "Cost multiplier")
		fs.Parse(args[1:])

		if *provider == "" || *model == "" {
			fmt.Println("Provider and model are required")
			os.Exit(1)
		}

		err := database.EnableModel(*provider, *model, *multiplier)
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
		} else {
			fmt.Printf("Model %s enabled on provider %s (Multiplier: %.2f)\n", *model, *provider, *multiplier)
			fmt.Println("Syncing pricing data from Oracle...")
			if err := workers.SyncPricing(database); err != nil {
				fmt.Printf("Warning: Failed to sync pricing: %v\n", err)
			}
		}

	case "disable":
		fs := flag.NewFlagSet("models disable", flag.ExitOnError)
		model := fs.String("model", "", "Model name")
		fs.Parse(args[1:])

		err := database.DisableModel(*model)
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
		} else {
			fmt.Println("Model disabled.")
		}

	case "list":
		models, _ := database.GetActiveModels()
		for _, m := range models {
			fmt.Printf("ID: %d | Provider: %s | Model: %s | Multiplier: %.2f | PromptCost: $%f | CompCost: $%f\n", m.ID, m.ProviderName, m.ModelName, m.CostMultiplier, m.PromptCost, m.CompletionCost)
		}

	case "sync-pricing":
		fmt.Println("Starting manual pricing sync from OpenRouter Oracle...")
		err := workers.SyncPricing(database)
		if err != nil {
			fmt.Printf("Sync failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Pricing sync finished successfully.")
	default:
		fmt.Printf("Unknown subcommand: %s\n", args[0])
	}
}
