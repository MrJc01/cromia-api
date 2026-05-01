package middleware

import (
	"context"
	"cromia/api/internal/db"
	"cromia/api/internal/security"
	"net/http"
	"strings"
	"sync"
	"time"
)

// contextKey é um tipo privado para evitar colisão de chaves no context.
type contextKey string

const APIKeyContextKey contextKey = "apiKey"

// Cache de autenticação para evitar Argon2 pesado em cada request
var authCache sync.Map

type cachedAuth struct {
	key    db.APIKey
	expiry time.Time
}

// AuthMiddleware valida o token Bearer contra os hashes Argon2id do banco de dados.
func AuthMiddleware(database db.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
			return
		}

		rawKey := strings.TrimPrefix(authHeader, "Bearer ")
		if rawKey == "" {
			http.Error(w, `{"error":"empty token"}`, http.StatusUnauthorized)
			return
		}

		// 1. Tenta buscar no cache primeiro
		if val, ok := authCache.Load(rawKey); ok {
			c := val.(cachedAuth)
			if time.Now().Before(c.expiry) {
				ctx := context.WithValue(r.Context(), APIKeyContextKey, &c.key)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			authCache.Delete(rawKey)
		}

		// 2. Busca no banco e valida via Argon2 (pesado)
		activeKeys, err := database.GetActiveKeys()
		if err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}

		var matchedKey *db.APIKey
		for i := range activeKeys {
			ok, err := security.CompareAPIKey(rawKey, activeKeys[i].KeyHash)
			if err == nil && ok {
				matchedKey = &activeKeys[i]
				break
			}
		}

		if matchedKey == nil {
			http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
			return
		}

		// 3. Salva no cache por 5 minutos
		authCache.Store(rawKey, cachedAuth{
			key:    *matchedKey,
			expiry: time.Now().Add(5 * time.Minute),
		})

		// Injeta a APIKey autenticada no context
		ctx := context.WithValue(r.Context(), APIKeyContextKey, matchedKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
