package main

import (
	"cromia/api/internal/config"
	"cromia/api/internal/db"
	"cromia/api/internal/handlers"
	"cromia/api/internal/middleware"
	"cromia/api/internal/python"
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg := config.Load()

	database, err := db.NewDB(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatalf("Não foi possível conectar ao banco de dados: %v", err)
	}

	pool := python.NewWorkerPool(cfg.PythonWorkers, 100)

	chatHandler := &handlers.ChatHandler{Pool: pool, DB: database}
	keyHandler := &handlers.KeyHandler{DB: database}
	modelsHandler := &handlers.ModelsHandler{DB: database}

	mux := http.NewServeMux()

	// Interceptor centralizado com middlewares de auth e rate limit
	// Para Chat Completions (/v1/chat/completions)
	chatPipe := middleware.AuthMiddleware(database,
		middleware.RateLimitMiddleware(
			http.HandlerFunc(chatHandler.Completions),
		),
	)
	mux.Handle("POST /v1/chat/completions", chatPipe)

	// Rota para Modelos
	modelsPipe := middleware.AuthMiddleware(database,
		middleware.RateLimitMiddleware(
			http.HandlerFunc(modelsHandler.List),
		),
	)
	mux.Handle("GET /v1/models", modelsPipe)

	// Rotas administrativas protegidas pela MASTER_API_KEY
	mux.HandleFunc("POST /v1/keys/generate", keyHandler.CreateKey)
	mux.HandleFunc("GET /v1/keys", keyHandler.ListKeys)
	mux.HandleFunc("DELETE /v1/keys/{id}", keyHandler.RevokeKey)

	// Endpoint de health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","version":"1.2.0"}`)
	})

	// Logger middleware global
	mainHandler := middleware.LoggingMiddleware(mux)

	fmt.Printf("Crom IA API iniciando na porta :%s\n", cfg.Port)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mainHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: time.Duration(cfg.PythonTimeoutSec+30) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
