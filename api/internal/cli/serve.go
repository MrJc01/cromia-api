package cli

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"cromia/api/internal/config"
	"cromia/api/internal/db"
	"cromia/api/internal/handlers"
	"cromia/api/internal/middleware"
)

func RunServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.String("port", "", "Port to run the unified server (default reads from .env PORT)")
	apiPort := fs.String("api-port", "", "Specific port for API (separates API from Web)")
	webPort := fs.String("web-port", "", "Specific port for Web (separates API from Web)")
	fs.Parse(args)

	cfg := config.Load()
	
	finalPort := cfg.Port
	if *port != "" {
		finalPort = *port
	}

	database, err := db.NewDB(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	chatHandler := &handlers.ChatHandler{DB: database}
	modelsHandler := &handlers.ModelsHandler{DB: database}
	balanceHandler := &handlers.BalanceHandler{DB: database}
	estimateHandler := &handlers.EstimateHandler{DB: database}
	usageHandler := &handlers.UsageHandler{DB: database}

	mux := http.NewServeMux()

	webHandler := &handlers.WebHandler{DB: database}

	// Web Dashboard Routes
	mux.HandleFunc("/", webHandler.ServeHome)
	mux.HandleFunc("/dashboard", webHandler.ServeDashboard)
	mux.HandleFunc("/login", webHandler.Login)
	mux.HandleFunc("/logout", webHandler.Logout)
	mux.HandleFunc("/v1/admin/me", webHandler.APIAdminMe)
	mux.HandleFunc("/v1/admin/usage", webHandler.APIAdminUsage)

	// API Routes
	chatPipe := middleware.AuthMiddleware(database,
		middleware.RateLimitMiddleware(
			middleware.BillingMiddleware(database,
				http.HandlerFunc(chatHandler.Completions),
			),
		),
	)
	mux.Handle("POST /v1/chat/completions", chatPipe)

	modelsPipe := middleware.AuthMiddleware(database,
		middleware.RateLimitMiddleware(
			http.HandlerFunc(modelsHandler.List),
		),
	)
	mux.Handle("GET /v1/models", modelsPipe)

	balancePipe := middleware.AuthMiddleware(database,
		middleware.RateLimitMiddleware(
			http.HandlerFunc(balanceHandler.GetBalance),
		),
	)
	mux.Handle("GET /v1/balance", balancePipe)

	estimatePipe := middleware.AuthMiddleware(database,
		middleware.RateLimitMiddleware(
			http.HandlerFunc(estimateHandler.EstimateCost),
		),
	)
	mux.Handle("POST /v1/estimate", estimatePipe)

	usagePipe := middleware.AuthMiddleware(database,
		middleware.RateLimitMiddleware(
			http.HandlerFunc(usageHandler.GetUsage),
		),
	)
	mux.Handle("GET /v1/usage", usagePipe)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","version":"1.2.0"}`)
	})

	mainHandler := middleware.LoggingMiddleware(mux)

	// Separation logic placeholder
	if *apiPort != "" && *webPort != "" {
		log.Printf("Starting separated... API on %s, Web on %s\n", *apiPort, *webPort)
		// We'll implement separated listeners if really needed, unified for now
	}

	fmt.Printf("CromIA Gateway Server running on :%s\n", finalPort)

	srv := &http.Server{
		Addr:         ":" + finalPort,
		Handler:      mainHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second, // Long timeout for SSE
		IdleTimeout:  120 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
