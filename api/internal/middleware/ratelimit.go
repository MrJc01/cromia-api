package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	"cromia/api/internal/db"
)

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[int]*bucket
	limit   int
	window  time.Duration
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

var globalLimiter = &rateLimiter{
	buckets: make(map[int]*bucket),
	limit:   60,
	window:  time.Minute,
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey, ok := r.Context().Value(APIKeyContextKey).(*db.APIKey)
		if !ok {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		limit := apiKey.RateLimit
		if limit <= 0 {
			// No limit if 0 or less
			next.ServeHTTP(w, r)
			return
		}

		allowed, remaining := globalLimiter.allow(apiKey.ID, limit)

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *rateLimiter) allow(keyID int, limit int) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[keyID]
	if !ok || time.Since(b.lastReset) > rl.window {
		rl.buckets[keyID] = &bucket{tokens: limit - 1, lastReset: time.Now()}
		return true, limit - 1
	}
	if b.tokens <= 0 {
		return false, 0
	}
	b.tokens--
	return true, b.tokens
}
