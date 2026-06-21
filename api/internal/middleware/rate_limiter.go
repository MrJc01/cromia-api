package middleware

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// IPRateLimiter armazena os limiters de cada IP
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter cria um IPRateLimiter.
// r = taxa (req/segundo), b = burst (lote maximo).
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	// Limpeza de IPs ociosos nao esta implementada nesta versao simples
	return i
}

// GetLimiter retorna o limiter para um IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}
	i.mu.Unlock()
	return limiter
}

// RateLimit middleware protege as rotas
func (i *IPRateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// Se falhar o parse (ex: test environment), usa o RemoteAddr direto
			ip = r.RemoteAddr
		}

		limiter := i.GetLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "429 Too Many Requests - Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
