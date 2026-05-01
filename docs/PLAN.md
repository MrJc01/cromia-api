---

🧠 Crom IA API — Plano Completo Final


---

1. 🎯 Objetivo

Construir uma API de IA compatível com OpenAI, com:

Backend em Go (alta performance)

Execução de modelos via Python embutido (GoPy)

Arquitetura segura, modular e escalável



---

2. 🏗️ Arquitetura Final

[ Client ]
    ↓
[ Go API Server ]
    ↓
[ Middleware (Auth, Rate Limit, Validation) ]
    ↓
[ Python Worker Pool (Go-managed) ]
    ↓
[ GoPy Binding Layer ]
    ↓
[ Python Runtime (embedded) ]
    ↓
[ Model Runner (PyTorch-ready) ]


---

3. 📦 Estrutura do Projeto

crom-ia/
│
├── api/ (Go)
│   ├── cmd/
│   │   └── server/
│   ├── internal/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   ├── services/
│   │   ├── python/        ← integração GoPy
│   │   ├── db/
│   │   ├── models/
│   │   ├── security/
│   │   └── config/
│   └── pkg/
│
├── worker/ (Python)
│   ├── crom/
│   │   ├── __init__.py
│   │   ├── runner.py      ← interface principal
│   │   ├── engine/
│   │   │   └── model_runner.py
│   │   └── schemas/
│
├── migrations/
├── seeds/
├── scripts/
└── docs/


---

4. 🔌 Endpoints (OpenAI Compatível — Texto)

4.1 POST /v1/chat/completions (principal)

4.2 POST /v1/completions

4.3 GET /v1/models

4.4 GET /v1/models/{id}


---

5. 🔐 Segurança (CORE DO SISTEMA)

5.1 API Keys

Prefixo: crom_sk_

Armazenamento: Argon2id hash


api_keys:
- id
- name
- key_hash
- created_at
- revoked_at


---

5.2 Autenticação

Header:

Authorization: Bearer crom_sk_xxx

Fluxo:

1. Extrair token


2. Hash com Argon2 verify


3. Validar revogação




---

5.3 Rate Limiting

Por API Key

Implementação inicial:

In-memory (map + mutex)


Futuro:

Redis




---

5.4 Proteções adicionais

Limite de payload (ex: 1MB)

Timeout global

Sanitização de input

Logs sem dados sensíveis



---

6. 🗄️ Banco de Dados

6.1 Suporte

SQLite (dev)

PostgreSQL (prod)



---

6.2 Tabelas

api_keys

requests

models


---

6.3 Abstração

type DB interface {
    GetAPIKeyByHash(...)
    CreateRequest(...)
    ListModels(...)
}


---

7. 🌱 Seed Database

Script obrigatório:

Criar modelo crom-1

Inserir API Key inicial


Fluxo seguro:

1. Gerar key offline


2. Hash com Argon2


3. Inserir no DB



🚫 Sem endpoint de criação de API Key


---

8. 🐍 Integração Go ↔ Python (GoPy)


---

8.1 Princípio

Sem HTTP. Comunicação direta via binding.

Go → GoPy → Python


---

8.2 Contrato Python (FIXO)

def run_model(model: str, messages: list, params: dict) -> dict:
    return {
        "output": "text",
        "usage": {
            "prompt_tokens": 0,
            "completion_tokens": 0
        }
    }


---

8.3 Binding

Gerado com:

gopy build ./worker/crom


---

8.4 Uso em Go

result := crom.RunModel(...)


---

9. 🧵 Worker Pool (CRÍTICO)

Problema: GIL (Python só executa 1 thread por vez)


---

Solução: Pool controlado pelo Go

Requests → Queue → Workers → Python


---

9.1 Estrutura

type Job struct {
    Input  Request
    Result chan Response
}

type Worker struct {
    id int
    jobs chan Job
}


---

9.2 Funcionamento

N workers (ex: 2–4)

Cada worker:

Inicializa Python runtime

Executa jobs sequencialmente


Sem concorrência dentro do Python



---

9.3 Benefícios

Evita conflito com GIL

Controle de carga

Previsibilidade



---

10. 🧨 Isolamento e Estabilidade

Risco:

Crash no Python → derruba processo


---

Mitigações

Panic recovery no Go

Timeout por execução

Pool reinicializável



---

Futuro:

Multi-processo Go

Supervisão externa (Docker/K8s)



---

11. 🐍 Worker Python (Preparado para PyTorch)

Estrutura

runner.py → entrada
model_runner.py → execução


---

Implementação inicial

Sem PyTorch

Retorno mockado



---

Futuro

Carregamento de modelos

GPU support

batching



---

12. ⚙️ Configuração

.env

API_PORT=8080

DB_TYPE=postgres
DB_DSN=...

PYTHON_WORKERS=2
PYTHON_TIMEOUT_MS=5000


---

13. 🔄 Fluxo Completo

1. Request chega
2. Auth (Argon2)
3. Rate limit
4. Validação
5. Entra na fila
6. Worker pega job
7. Executa via GoPy
8. Retorna resultado
9. Salva no DB
10. Responde cliente (OpenAI format)


---

14. 🧩 Middleware (Go)

Auth

Rate limit

Logging

Recovery

Timeout



---

15. 📊 Observabilidade

Logs JSON estruturados

Métricas (futuro)

Tracing (futuro)



---

16. 🚀 Deploy

Dev

Binário Go + Python instalado


Produção

Docker:

Go + Python no mesmo container



⚠️ Fixar versão do Python (ex: 3.11)


---

17. ⚠️ Decisões Críticas

✔ GoPy ao invés de HTTP
✔ Worker pool obrigatório
✔ Sem criação de API Key via API
✔ Python isolado logicamente
✔ Segurança antes de performance


---

18. 🔮 Roadmap Futuro

Streaming (stream: true)

Embeddings

Cache de respostas

Multi-model routing

GPU workers

Migração para gRPC (se necessário)



---

19. 🧠 Resumo Final

A Crom IA API será:

⚡ Extremamente rápida (sem rede)

🔐 Segura (Argon2 + validação)

🧵 Controlada (worker pool vs GIL)

🧩 Modular (Go + Python desacoplados logicamente)

🔄 Compatível com OpenAI
