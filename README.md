# CromIA API Gateway

A **CromIA API** é um poderoso API Gateway focado em faturamento (billing), compatível com o formato OpenAI. Ela atua como um roteador e proxy para provedores de LLM como **DeepSeek** e **OpenRouter**, descontando créditos dos usuários em tempo real baseado no consumo exato de tokens.

Todo o projeto é construído em Go, compilado para um único binário que contém o servidor API, a interface de linha de comando (CLI) para administração e o Dashboard Web.

---

## 🚀 Como Executar

### 1. Compilar o Binário
```bash
go build -o cromia api/cmd/server/main.go
```

### 2. Configurar Variáveis de Ambiente (`.env`)
```env
# Porta padrão para o comando "serve"
PORT=8080

# Chaves de Integração (Provedores)
DEEPSEEK_API_KEY=sk-suachavedeepseek
OPENROUTER_API_KEY=sk-or-v1-suachaveopenrouter
```

### 3. Iniciar o Servidor
```bash
./cromia serve
```

---

## 🛠️ Interface de Linha de Comando (CLI)

O binário `cromia` atua como ferramenta de gestão local. **Não é necessário editar o banco de dados manualmente.**

- `cromia users create --username joao --password senha123`
- `cromia users add-credits --user joao --amount 5000`
- `cromia keys generate --user joao --name "App Principal"`
- `cromia models enable --provider deepseek --model deepseek-chat --multiplier 1.5`

Para mais comandos e detalhes, verifique a documentação do CLI: [docs/CLI.md](docs/CLI.md).

---

## 🌐 Dashboard Web Embutido

Acesse `http://localhost:8080/` para visualizar o Dashboard Web embutido.
Os usuários finais podem fazer login com as credenciais criadas pelo administrador via CLI, ver seu saldo e histórico de uso.

---

## 📚 Documentação Adicional

Acesse a pasta `docs/` para ver as especificações arquiteturais:
- [Arquitetura](docs/ARCHITECTURE.md)
- [Banco de Dados](docs/DATABASE.md)
- [Sistema de Faturamento](docs/BILLING.md)
- [Referência do CLI](docs/CLI.md)
- [Interface Web](docs/WEB.md)
