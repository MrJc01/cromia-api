package cli

import (
	"fmt"
	"os"
)

func PrintUsage() {
	fmt.Println("CromIA API Gateway CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  cromia <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  serve       Start the HTTP server (API & Web)")
	fmt.Println("  users       Manage users and credits")
	fmt.Println("  keys        Manage API keys")
	fmt.Println("  models      Manage provider models")
	fmt.Println("  monitor     Monitor real-time usage (coming soon)")
	fmt.Println("\nUse \"cromia <command> --help\" for more information about a command.")
}

func Execute() {
	if len(os.Args) < 2 {
		PrintUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		RunServe(os.Args[2:])
	case "users":
		RunUsers(os.Args[2:])
	case "keys":
		RunKeys(os.Args[2:])
	case "models":
		RunModels(os.Args[2:])
	case "monitor":
		RunMonitor(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		PrintUsage()
		os.Exit(1)
	}
}
