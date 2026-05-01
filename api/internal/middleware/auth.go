package middleware

import (
	"context"
	"fmt"
	"cromia/api/internal/db"
	"cromia/api/internal/security"
	"net/http"
	"strings"
	"sync"
	"time"
)

type contextKey string

const APIKeyContextKey contextKey = "apiKey"
const UserContextKey contextKey = "user"

var authCache sync.Map

type cachedAuth struct {
	key    db.APIKey
	user   db.User
	expiry time.Time
}

func AuthMiddleware(database db.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
			return
		}

		rawKey := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if rawKey == "" {
			http.Error(w, `{"error":"empty token"}`, http.StatusUnauthorized)
			return
		}

		if val, ok := authCache.Load(rawKey); ok {
			c := val.(cachedAuth)
			if time.Now().Before(c.expiry) {
				ctx := context.WithValue(r.Context(), APIKeyContextKey, &c.key)
				ctx = context.WithValue(ctx, UserContextKey, &c.user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			authCache.Delete(rawKey)
		}

		activeKeys, err := database.GetActiveKeys()
		if err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
		fmt.Printf("Found %d active keys\n", len(activeKeys))

		var matchedKey *db.APIKey
		for i := range activeKeys {
			ok, err := security.CompareAPIKey(rawKey, activeKeys[i].KeyHash)
			if err != nil {
				fmt.Printf("Compare error: %v (rawKey len: %d, hash len: %d)\n", err, len(rawKey), len(activeKeys[i].KeyHash))
			}
			if err == nil && ok {
				matchedKey = &activeKeys[i]
				break
			}
		}

		if matchedKey == nil {
			http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
			return
		}

		user, err := database.GetUserByID(matchedKey.UserID)
		if err != nil || user == nil {
			http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
			return
		}

		authCache.Store(rawKey, cachedAuth{
			key:    *matchedKey,
			user:   *user,
			expiry: time.Now().Add(1 * time.Minute),
		})

		ctx := context.WithValue(r.Context(), APIKeyContextKey, matchedKey)
		ctx = context.WithValue(ctx, UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
