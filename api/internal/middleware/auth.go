package middleware

import (
	"context"
	"cromia/api/internal/db"
	"cromia/api/internal/security"
	"cromia/api/internal/utils"
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
			utils.JSONError(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		rawKey := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if rawKey == "" {
			utils.JSONError(w, "empty token", http.StatusUnauthorized)
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

		hashedKey, _ := security.HashAPIKey(rawKey)
		matchedKey, err := database.GetKeyByHash(hashedKey)
		if err != nil {
			utils.JSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if matchedKey == nil {
			utils.JSONError(w, "invalid API key", http.StatusUnauthorized)
			return
		}

		user, err := database.GetUserByID(matchedKey.UserID)
		if err != nil || user == nil {
			utils.JSONError(w, "user not found", http.StatusUnauthorized)
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
