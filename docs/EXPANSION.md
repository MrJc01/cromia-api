# 🚀 Crom IA API — Plano de Expansão Completo

> Documento gerado após leitura integral de todos os arquivos do projeto.  
> Estado analisado: `api/`, `worker/`, `go.mod`, `Makefile`, `.env`, `PLAN.md`, `README.md`

---

## 1. 📊 Estado Atual (O que já existe)

### ✅ Implementado
| Componente | Arquivo | Status |
|---|---|---|
| Servidor HTTP Go (net/http) | `api/cmd/server/main.go` | ✅ Funcional |
| Auth middleware (Argon2id) | `api/internal/middleware/auth.go` | ✅ Funcional |
| Chat handler (normal + SSE) | `api/internal/handlers/chat.go` | ✅ Funcional |
| Geração de API Keys | `api/internal/handlers/key_handler.go` | ✅ Funcional |
| DB abstrato (SQLite/Postgres) | `api/internal/db/db.go` | ✅ Funcional |
| Security (Argon2id, random) | `api/internal/security/security.go` | ✅ Funcional |
| Worker Pool (Unix Socket) | `api/internal/python/worker_pool.go` | ✅ Funcional |
| Socket Server Python | `worker/crom/socket_server.py` | ✅ Funcional |
| Model Runner (llama-cpp + fallback) | `worker/crom/engine/model_runner.py` | ✅ Funcional |
| Restart automático de workers | `worker_pool.go#supervise()` | ✅ Funcional |
| Streaming SSE | `chat.go` + `socket_server.py` | ✅ Funcional |
| Auto-migrate das tabelas | `db.go#migrate()` | ✅ Funcional |

### ❌ Lacunas Identificadas no Código
| Problema | Local | Impacto |
|---|---|---|
| **Rate limiting ausente** | `main.go` não registra middleware de rate limit | 🔴 Crítico |
| `GetAPIKeyByHash` retorna sempre `sql.ErrNoRows` | `db.go:90` — stub não implementado | 🟡 Médio |
| `CreateRequest` nunca é chamado | `chat.go` não registra requests no DB | 🟡 Médio |
| `GET /v1/models` não existe | Planejado no PLAN.md mas não implementado | 🟡 Médio |
| `POST /v1/completions` não existe | Planejado no PLAN.md mas não implementado | 🟡 Médio |
| `DELETE /v1/keys/{id}` (revogação) não existe | `revoked_at` existe no DB, mas sem endpoint | 🟡 Médio |
| Auth da rota `/v1/keys/generate` é string-compare direto | `key_handler.go:23` — vulnerável a timing attack | 🟠 Baixo |
| Payload sem limite de tamanho | `chat.go` lê body sem `http.MaxBytesReader` | 🟠 Baixo |
| Timeout por request não configurado | `main.go` usa `http.ListenAndServe` puro | 🟠 Baixo |
| Logs sem estrutura JSON | `log.Printf` simples | 🟢 Melhoria |
| `PYTHON_WORKERS` hardcoded como 2 em `main.go` | `main.go:31` ignora a env var | 🟡 Médio |

---

## 2. 🔒 Rate Limiting — Implementação Detalhada

### 2.1 Fase 1: In-Memory (imediato)

Criar `api/internal/middleware/ratelimit.go`:

```go
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
        apiKey := r.Context().Value(APIKeyContextKey).(*db.APIKey)

        allowed, remaining := globalLimiter.allow(apiKey.ID)

        w.Header().Set("X-RateLimit-Limit", "60")
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

func (rl *rateLimiter) allow(keyID int) (bool, int) {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    b, ok := rl.buckets[keyID]
    if !ok || time.Since(b.lastReset) > rl.window {
        rl.buckets[keyID] = &bucket{tokens: rl.limit - 1, lastReset: time.Now()}
        return true, rl.limit - 1
    }
    if b.tokens <= 0 {
        return false, 0
    }
    b.tokens--
    return true, b.tokens
}
```

**Registrar no `main.go`** encadeando após AuthMiddleware:
```go
mux.Handle(
    "/v1/chat/completions",
    middleware.AuthMiddleware(database,
        middleware.RateLimitMiddleware(
            http.HandlerFunc(chatHandler.Completions),
        ),
    ),
)
```

**Headers de resposta:**
- `X-RateLimit-Limit: 60`
- `X-RateLimit-Remaining: N`
- `X-RateLimit-Reset: <unix timestamp>`
- `Retry-After: 60` (em respostas 429)

### 2.2 Fase 2: Rate Limits Configuráveis por Key

Adicionar coluna no banco:
```sql
ALTER TABLE api_keys ADD COLUMN rate_limit INTEGER NOT NULL DEFAULT 60;
ALTER TABLE api_keys ADD COLUMN rate_limit_window TEXT NOT NULL DEFAULT '1m';
```

Atualizar o struct `APIKey`:
```go
type APIKey struct {
    ID              int
    Name            string
    KeyHash         string
    CreatedAt       string
    RevokedAt       string
    RateLimit       int    // requests por janela (0 = sem limite)
    RateLimitWindow string // "1m", "1h", "1d"
}
```

### 2.3 Fase 3: Redis (Produção Distribuída)

```go
import "github.com/redis/go-redis/v9"

func (rl *redisRateLimiter) allow(ctx context.Context, keyID int) (bool, int) {
    key := fmt.Sprintf("rl:key:%d", keyID)
    pipe := rl.client.Pipeline()
    incr := pipe.Incr(ctx, key)
    pipe.Expire(ctx, key, rl.window)
    pipe.Exec(ctx)

    count := int(incr.Val())
    remaining := rl.limit - count
    if remaining < 0 {
        remaining = 0
    }
    return count <= rl.limit, remaining
}
```

**Variáveis de ambiente a adicionar:**
```env
RATE_LIMIT_BACKEND=memory     # ou "redis"
REDIS_URL=redis://localhost:6379
RATE_LIMIT_PER_KEY=60
RATE_LIMIT_WINDOW=1m
```

---

## 3. 🔧 Correções Urgentes (Quick Wins)

### 3.1 Ler PYTHON_WORKERS da env var
**Arquivo:** `api/cmd/server/main.go`
```go
// Antes (linha 31):
numWorkers := 2

// Depois:
numWorkers := 2
if v := os.Getenv("PYTHON_WORKERS"); v != "" {
    if n, err := strconv.Atoi(v); err == nil && n > 0 {
        numWorkers = n
    }
}
```

### 3.2 Limite de payload no chat handler
**Arquivo:** `api/internal/handlers/chat.go`
```go
// Adicionar no início de Completions():
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
```

### 3.3 Timeout global no servidor HTTP
**Arquivo:** `api/cmd/server/main.go`
```go
srv := &http.Server{
    Addr:         ":" + port,
    Handler:      mux,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 90 * time.Second, // maior por causa de streaming
    IdleTimeout:  120 * time.Second,
}
log.Fatal(srv.ListenAndServe())
```

### 3.4 Registrar requests no DB
**Arquivo:** `api/internal/handlers/chat.go`
```go
// Após obter result com sucesso:
apiKey := r.Context().Value(middleware.APIKeyContextKey).(*db.APIKey)
go h.DB.CreateRequest(apiKey.ID, req.Model, lastUserMessage(req.Messages))
```

### 3.5 Timing-safe compare na Master Key
**Arquivo:** `api/internal/handlers/key_handler.go`
```go
// Antes (vulnerável a timing attack):
if authHeader != "Bearer "+masterKey { ... }

// Depois (constant time):
import "crypto/subtle"
expected := []byte("Bearer " + masterKey)
actual   := []byte(authHeader)
if subtle.ConstantTimeCompare(expected, actual) != 1 { ... }
```

---

## 4. 📡 Endpoints Faltantes

### 4.1 GET /v1/models
**Arquivo novo:** `api/internal/handlers/models_handler.go`

Resposta no formato OpenAI:
```json
{
  "object": "list",
  "data": [
    {
      "id": "crom-1",
      "object": "model",
      "created": 1714000000,
      "owned_by": "cromia"
    }
  ]
}
```

### 4.2 GET /v1/models/{id}
Retorna detalhes de um modelo específico ou 404.

### 4.3 POST /v1/completions
Endpoint legado (texto simples, sem chat). Mapeia `prompt` string para `messages`.

### 4.4 DELETE /v1/keys/{id} — Revogação
```go
// Adicionar ao DB interface:
RevokeKey(id int) error

// SQL:
// UPDATE api_keys SET revoked_at = CURRENT_TIMESTAMP WHERE id = ?
```

### 4.5 GET /v1/keys — Listagem (admin)
Apenas para quem usa a `MASTER_API_KEY`. Retorna lista de keys ativas com `name`, `id`, `created_at` (nunca o hash).

### 4.6 GET /health
```go
mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "ok",
        "version": "1.0.0",
    })
})
```

---

## 5. 🧠 Cache de Respostas

### 5.1 Cache em memória
```go
// api/internal/cache/cache.go
type ResponseCache struct {
    mu    sync.RWMutex
    store map[string]cachedEntry
    ttl   time.Duration
}

type cachedEntry struct {
    response  python.Response
    expiresAt time.Time
}

func (c *ResponseCache) key(model string, messages []map[string]string) string {
    h := sha256.New()
    json.NewEncoder(h).Encode(model)
    json.NewEncoder(h).Encode(messages)
    return hex.EncodeToString(h.Sum(nil))
}
```

> Cache só se aplica a requests sem `stream: true`. Adicionar header `X-Cache: HIT` ou `MISS`.

### 5.2 Cache Redis (produção)
TTL configurável por modelo. Usar `SET key value EX ttl NX` para atomicidade.

---

## 6. 📊 Observabilidade

### 6.1 Structured Logging com slog (Go 1.21+)
```go
import "log/slog"

slog.Info("request completed",
    "method", r.Method,
    "path", r.URL.Path,
    "status", statusCode,
    "latency_ms", latency.Milliseconds(),
    "api_key_id", apiKey.ID,
    "model", req.Model,
)
```

### 6.2 Logging Middleware
**Arquivo:** `api/internal/middleware/logger.go`
```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        rw := &statusRecorder{ResponseWriter: w, status: 200}
        next.ServeHTTP(rw, r)
        slog.Info("http",
            "method", r.Method,
            "path", r.URL.Path,
            "status", rw.status,
            "latency_ms", time.Since(start).Milliseconds(),
            "remote", r.RemoteAddr,
        )
    })
}
```

### 6.3 Request ID Middleware
Gerar UUID por request e propagar em logs + response header `X-Request-ID`.

### 6.4 Métricas Prometheus (Fase futura)
Expor em `GET /metrics`:
- `cromia_requests_total{method, path, status}`
- `cromia_request_duration_seconds{model}`
- `cromia_worker_queue_depth`
- `cromia_worker_restarts_total`

---

## 7. 🐍 Melhorias no Worker Python

### 7.1 GPU Support
```python
_llama_model_cache[model_name] = Llama(
    model_path=model_path,
    n_ctx=int(os.environ.get("CROM_CTX", "2048")),
    n_threads=int(os.environ.get("CROM_THREADS", "4")),
    n_gpu_layers=int(os.environ.get("CROM_GPU_LAYERS", "0")),  # <- novo
    verbose=False,
)
```

### 7.2 Parâmetros via env
| Env Var | Padrão | Descrição |
|---|---|---|
| `CROM_MODEL_PATH` | `models/crom-1.gguf` | Caminho do arquivo do modelo |
| `CROM_CTX` | `2048` | Context window em tokens |
| `CROM_THREADS` | `4` | Threads de inferência CPU |
| `CROM_MAX_TOKENS` | `512` | Máximo de tokens gerados |
| `CROM_TEMPERATURE` | `0.7` | Temperatura padrão |
| `CROM_GPU_LAYERS` | `0` | Camadas na GPU (0 = CPU only) |

### 7.3 Múltiplos modelos GGUF
```python
model_path = os.path.join(
    os.environ.get("CROM_MODELS_DIR", "models"),
    f"{model_name}.gguf"
)
```

### 7.4 Validação de input no socket_server.py
```python
def validate_request(request: dict) -> str | None:
    if not request.get("model"):
        return "missing 'model'"
    if not isinstance(request.get("messages"), list):
        return "missing or invalid 'messages'"
    if len(request["messages"]) > 100:
        return "too many messages (max 100)"
    return None
```

---

## 8. 🗄️ Banco de Dados

### 8.1 Índices faltantes — CRÍTICO
```sql
-- Auth hoje faz full-scan em api_keys a cada request!
CREATE INDEX IF NOT EXISTS idx_api_keys_active
    ON api_keys(revoked_at) WHERE revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_requests_api_key ON requests(api_key_id);
CREATE INDEX IF NOT EXISTS idx_requests_created ON requests(created_at);
```

### 8.2 Usage Tracking completo
```sql
ALTER TABLE requests ADD COLUMN prompt_tokens     INTEGER DEFAULT 0;
ALTER TABLE requests ADD COLUMN completion_tokens INTEGER DEFAULT 0;
ALTER TABLE requests ADD COLUMN latency_ms        INTEGER DEFAULT 0;
ALTER TABLE requests ADD COLUMN stream            BOOLEAN DEFAULT FALSE;
```

### 8.3 Migration System versionado
```
migrations/
  001_initial.sql
  002_add_rate_limit_columns.sql
  003_add_usage_tracking.sql
  004_add_key_expiration.sql
```

### 8.4 Compatibilidade SQLite/PostgreSQL
As queries atuais usam `?` (SQLite). PostgreSQL requer `$1, $2`.
Usar `sqlx` com `Rebind()` ou `squirrel` query builder para portabilidade.

---

## 9. 🏗️ Refatoração de Arquitetura

### 9.1 Config struct centralizada
**Arquivo:** `api/internal/config/config.go`
```go
type Config struct {
    Port             string
    DBDriver         string
    DBDSN            string
    MasterAPIKey     string
    PythonWorkers    int
    PythonTimeoutSec int
    RateLimitBackend string        // "memory" | "redis"
    RateLimitPerKey  int
    RateLimitWindow  time.Duration
    RedisURL         string
    LogLevel         string        // "debug" | "info" | "warn"
    CacheEnabled     bool
    CacheTTL         time.Duration
}
```

### 9.2 Router chi (para path params limpos)
```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()
r.Use(middleware.RequestID)
r.Use(LoggingMiddleware)
r.Use(middleware.Recoverer)
r.Get("/health", healthHandler)

r.Group(func(r chi.Router) {
    r.Use(AuthMiddleware(db))
    r.Use(RateLimitMiddleware(limiter))
    r.Post("/v1/chat/completions", chatHandler.Completions)
    r.Post("/v1/completions", chatHandler.CompletionsLegacy)
    r.Get("/v1/models", modelsHandler.List)
    r.Get("/v1/models/{id}", modelsHandler.Get)
})

r.Group(func(r chi.Router) {
    r.Use(MasterKeyMiddleware)
    r.Post("/v1/keys/generate", keyHandler.Create)
    r.Get("/v1/keys", keyHandler.List)
    r.Delete("/v1/keys/{id}", keyHandler.Revoke)
})
```

### 9.3 Graceful Shutdown
```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Shutdown(ctx)
pool.Shutdown() // drena jobs antes de matar workers Python
```

---

## 10. 🚀 Deploy

### 10.1 Dockerfile (multi-stage)
```dockerfile
FROM golang:1.25-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o bin/cromia-api api/cmd/server/main.go

FROM python:3.11-slim
WORKDIR /app
RUN pip install --no-cache-dir llama-cpp-python
COPY --from=builder /app/bin/cromia-api .
COPY worker/ ./worker/
EXPOSE 8080
HEALTHCHECK --interval=30s CMD curl -f http://localhost:8080/health || exit 1
CMD ["./cromia-api"]
```

### 10.2 docker-compose.yml
```yaml
services:
  api:
    build: .
    ports: ["8080:8080"]
    env_file: .env
    depends_on: [postgres, redis]
    volumes:
      - ./models:/app/models

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: cromia
      POSTGRES_USER: cromia
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    command: redis-server --maxmemory 256mb --maxmemory-policy allkeys-lru

volumes:
  pgdata:
```

---

## 11. 🗺️ Roadmap por Fases

### 🔴 Fase 1 — Correções Críticas (1–2 dias)
- [ ] Rate limiting in-memory por API Key (`middleware/ratelimit.go`)
- [ ] Ler `PYTHON_WORKERS` do env (corrigir hardcode em `main.go:31`)
- [ ] `http.MaxBytesReader` no chat handler (1MB)
- [ ] `ReadTimeout` / `WriteTimeout` / `IdleTimeout` no servidor HTTP
- [ ] Timing-safe compare na Master Key (`subtle.ConstantTimeCompare`)
- [ ] Chamar `CreateRequest` no DB após inferência bem-sucedida
- [ ] Índice `idx_api_keys_active` no banco (elimina full-scan no auth)
- [ ] `GET /health` endpoint

### 🟡 Fase 2 — Features e Endpoints (3–5 dias)
- [ ] `GET /v1/models` e `GET /v1/models/{id}`
- [ ] `POST /v1/completions` (legado)
- [ ] `DELETE /v1/keys/{id}` (revogação)
- [ ] `GET /v1/keys` (listagem admin)
- [ ] Rate limit configurável por key (coluna no DB)
- [ ] Structured logging com `slog`
- [ ] Logging middleware com request ID
- [ ] Config struct centralizada (`internal/config`)
- [ ] Validação de input completa (model, messages count, content length)

### 🟢 Fase 3 — Performance e Observabilidade (1–2 semanas)
- [ ] Cache de respostas in-memory (SHA256 do payload)
- [ ] Migrar para router `chi`
- [ ] Sistema de migrations versionado
- [ ] Métricas Prometheus em `/metrics`
- [ ] Compatibilidade PostgreSQL via sqlx
- [ ] Graceful shutdown com drain do worker pool
- [ ] CORS middleware
- [ ] Usage tracking completo (tokens, latência no DB)

### 🔵 Fase 4 — Escalabilidade (2–4 semanas)
- [ ] Rate limiting via Redis
- [ ] Cache Redis com TTL configurável
- [ ] Múltiplos modelos GGUF por `CROM_MODELS_DIR`
- [ ] GPU layers configurável (`CROM_GPU_LAYERS`)
- [ ] Batching de inferência no Python
- [ ] Dockerfile + docker-compose completo
- [ ] CI/CD (GitHub Actions)

### ⚪ Fase 5 — Produção Avançada (Futuro)
- [ ] API Key com expiração (`expires_at`)
- [ ] Dashboard de uso por key (tokens consumidos, requests/dia)
- [ ] Multi-model routing
- [ ] `POST /v1/embeddings`
- [ ] SDK Python e Node.js
- [ ] Documentação OpenAPI/Swagger
- [ ] Kubernetes manifests

---

## 12. 📁 Estrutura Final Proposta

```
cromia-api/
├── api/
│   ├── cmd/server/main.go          ← corrigir PYTHON_WORKERS, timeouts
│   └── internal/
│       ├── cache/                  ← NOVO: cache.go
│       ├── config/                 ← NOVO: config.go
│       ├── db/
│       │   └── db.go               ← EXPANDIR: índices, usage tracking
│       ├── handlers/
│       │   ├── chat.go             ← EXPANDIR: MaxBytesReader, CreateRequest
│       │   ├── key_handler.go      ← CORRIGIR: ConstantTimeCompare
│       │   └── models_handler.go   ← NOVO
│       ├── middleware/
│       │   ├── auth.go
│       │   ├── ratelimit.go        ← NOVO
│       │   ├── logger.go           ← NOVO
│       │   └── cors.go             ← NOVO
│       ├── python/
│       │   └── worker_pool.go
│       └── security/
│           └── security.go
├── migrations/                     ← NOVO
│   ├── 001_initial.sql
│   └── 002_add_rate_limit.sql
├── worker/
│   └── crom/
│       ├── socket_server.py        ← EXPANDIR: validação de input
│       ├── runner.py
│       └── engine/
│           └── model_runner.py     ← EXPANDIR: GPU, multi-model, env vars
├── models/                         ← NOVO: pasta para arquivos .gguf
├── docker/                         ← NOVO
│   ├── Dockerfile
│   └── docker-compose.yml
├── .env
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

> **Prioridade máxima para produção:** Rate limiting + `MaxBytesReader` + timeouts HTTP.
> Esses três itens protegem contra DoS/abusos antes de qualquer feature nova.

> **ATENÇÃO:** O auth hoje faz **full-scan em `api_keys`** a cada request (compara Argon2
> para TODAS as chaves ativas). Com muitas chaves isso se torna lento. Criar o índice
> e considerar um cache em memória do resultado Argon2 por TTL curto (ex: 30s).
