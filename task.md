# Checklist de Execução: CromIA API Gateway

- [ ] **Fase 1: Limpeza e Preparação**
  - [ ] Remover diretório `worker/` (código Python).
  - [ ] Remover `api/internal/python/` (integração GoPy).
  - [ ] Limpar `go.mod` de dependências antigas (se houver).
  - [ ] Atualizar `README.md` refletindo a nova arquitetura de Gateway.

- [ ] **Fase 2: Banco de Dados e Modelagem (Users & Billing)**
  - [ ] Criar schema `users` (id, username, password_hash, balance, created_at).
  - [ ] Atualizar schema `api_keys` (vincular ao user_id).
  - [ ] Criar schema `provider_models` (id, provider_name, model_name, cost_multiplier, is_active).
  - [ ] Criar schema `usage_logs` (id, user_id, api_key_id, prompt_tokens, completion_tokens, cost, created_at).
  - [ ] Implementar camada de acesso a dados `db/users.go` e `db/billing.go`.

- [ ] **Fase 3: Arquitetura CLI (Comandos Administrativos)**
  - [ ] Refatorar `main.go` para usar o pacote `flag` padrão suportando subcomandos.
  - [ ] Implementar comando `serve` (`--api-port 8080`, `--web-port 8081`).
  - [ ] Implementar comandos de `users` (`create`, `add-credits`, `list`).
  - [ ] Implementar comandos de `keys` (`generate --user`, `revoke`).
  - [ ] Implementar comandos de `models` (`sync`, `enable`, `disable`).

- [ ] **Fase 4: Core da API Gateway (Proxy & Streaming)**
  - [ ] Implementar middleware de Autenticação baseado na nova tabela `api_keys`.
  - [ ] Implementar middleware de Validação de Saldo (Billing Check).
  - [ ] Criar módulo `providers/` com adapters para DeepSeek (`api.deepseek.com`) e OpenRouter (`openrouter.ai`).
  - [ ] Refatorar `handlers/chat.go` para atuar como proxy transparente (SSE byte-a-byte).
  - [ ] Implementar o listener pós-requisição que extrai o bloco `"usage"` do SSE e chama a dedução no DB.

- [ ] **Fase 5: Frontend Dashboard & Roteamento**
  - [ ] Criar diretório `web/static/` para HTML e assets.
  - [ ] Configurar `go:embed` para empacotar o diretório `web` no binário final.
  - [ ] Criar `index.html` minimalista (Landing page com Tailwind CSS).
  - [ ] Criar `dashboard.html` (Painel com Vanilla JS para login e vizualização de saldo/chaves).
  - [ ] Criar rotas no backend para servir o painel Web (`/login`, `/dashboard`).
  - [ ] Criar API interna `/v1/admin/me` (autenticada por cookie de sessão) para o dashboard buscar os dados do usuário.

- [ ] **Fase 6: Testes e Estabilização**
  - [ ] Testar fluxo completo: CLI User Create -> CLI Add Credits -> CLI Gen Key.
  - [ ] Testar consumo com chamada CURL para modelo DeepSeek.
  - [ ] Verificar dedução de créditos no banco e registro em `usage_logs`.
  - [ ] Testar acesso ao Dashboard via navegador (login e verificação de saldo).
