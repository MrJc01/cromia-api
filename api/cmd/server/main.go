package main

import (
	"cromia/api/internal/cli"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load() // Ignore error if .env doesn't exist
	cli.Execute()
}
