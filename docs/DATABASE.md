# Modelagem de Dados e Entidades

O banco de dados usa SQLite para facilitar o deploy sem dependências extras, mantendo persistência no arquivo `data.db`.

## 1. Tabelas Principais

### `users`
Tabela fundamental. Os usuários são os donos das contas no painel e detêm o saldo (créditos).
- `id` (INTEGER, PK)
- `username` (TEXT, Único) - Serve como login no painel.
- `password_hash` (TEXT) - Hash bcrypt/argon2 para login no dashboard.
- `balance` (DECIMAL) - Saldo atual de créditos (Ex: `1000.50`).
- `created_at` (DATETIME)

### `api_keys`
Chaves geradas para o usuário usar a API.
- `id` (INTEGER, PK)
- `user_id` (INTEGER, FK para users)
- `name` (TEXT) - Ex: "Chave do App de Teste".
- `key_hash` (TEXT) - Hash seguro (Argon2). A chave real só é vista uma vez na criação.
- `created_at` (DATETIME)
- `revoked_at` (DATETIME)

### `provider_models`
Catálogo de modelos ativados no Gateway.
- `id` (INTEGER, PK)
- `provider_name` (TEXT) - Ex: `deepseek` ou `openrouter`.
- `model_name` (TEXT) - Nome original da API (ex: `deepseek-chat`).
- `cost_multiplier` (DECIMAL) - Fator multiplicador de custo, caso a administração deseje cobrar mais que o provedor original (Ex: `1.5` cobra 50% de lucro).
- `is_active` (BOOLEAN) - Se `false`, a API rejeita chamadas para este modelo.

### `usage_logs`
Tabela de auditoria. Extremamente importante para resolver disputas de cobrança.
- `id` (INTEGER, PK)
- `user_id` (INTEGER, FK)
- `api_key_id` (INTEGER, FK)
- `model_used` (TEXT)
- `prompt_tokens` (INTEGER)
- `completion_tokens` (INTEGER)
- `cost_deducted` (DECIMAL) - Custo final deduzido do saldo.
- `created_at` (DATETIME)

## 2. Índices

Para alta performance no _Proxying_ e Autenticação:
- Índice em `api_keys(key_hash)` onde `revoked_at IS NULL`.
- Índice em `usage_logs(user_id)` para exibir relatórios rápidos no Dashboard.
