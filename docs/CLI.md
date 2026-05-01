# CLI da CromIA (Command Line Interface)

A principal forma de administrar o Gateway é usando o próprio binário da API via comandos do terminal. Isso elimina a necessidade de scripts Python ou painéis complexos de administração.

A estrutura é construída na forma `cromia <comando> <subcomando> [flags]`.

## Comandos Principais

### `serve`
Inicia a aplicação.
- `cromia serve` (Sobe a API e Web na mesma porta lida do `.env`).
- `cromia serve --port 8080` (Força a porta única).
- `cromia serve --api-port 8080 --web-port 8081` (Separa serviços em portas distintas).

---

### `users` (Gerenciamento de Contas)

- **Criar Conta:**
  `cromia users create --username joao --password senha123`
- **Adicionar Saldo (Créditos):**
  `cromia users add-credits --user joao --amount 5000`
- **Remover Saldo:**
  `cromia users remove-credits --user joao --amount 100`
- **Listar Usuários e Saldos:**
  `cromia users list`

---

### `keys` (Gerenciamento de API Keys)

- **Gerar Nova Chave:**
  `cromia keys generate --user joao --name "App Teste"`
  *(A chave é impressa no terminal apenas uma vez).*
- **Revogar Chave:**
  `cromia keys revoke --id 5`

---

### `models` (Gerenciamento de Provedores)

- **Habilitar/Adicionar Modelo:**
  `cromia models enable --provider deepseek --model deepseek-chat --multiplier 1.0`
- **Desabilitar Modelo:**
  `cromia models disable --model deepseek-chat`
- **Listar Modelos Ativos:**
  `cromia models list`

---

### `monitor` (Observabilidade)

- **Monitor em Tempo Real:**
  `cromia monitor`
  *(Mostra no console um streaming de logs formatados toda vez que uma requisição bate na API, mostrando `Usuário`, `Modelo`, `Latência`, `Custo`).*
