# Crom IA API - Documentação

A **Crom IA API** é uma API de inteligência artificial de alta performance, compatível com o formato da OpenAI, desenvolvida em Go com processamento de modelos via Python embutido.

---

## 🚀 Como Executar

1. **Configurar Variáveis de Ambiente (`.env`):**
   Crie um arquivo `.env` na raiz do projeto com as seguintes configurações:
   ```env
   MASTER_API_KEY=crom_sk_master_secret_123456
   DB_DRIVER=sqlite3
   DB_DSN=data.db
   PYTHON_WORKERS=2
   ```

2. **Iniciar o Servidor:**
   ```bash
   export $(grep -v '^#' .env | xargs)
   make run
   ```

---

## 🔐 Endpoints

### 1. Gerar API Key (Admin)
Utilize este endpoint para gerar chaves de acesso para consumidores da API.

- **URL:** `/v1/keys/generate`
- **Método:** `POST`
- **Headers:** `Authorization: Bearer <MASTER_API_KEY>`
- **Body:**
  ```json
  {
    "name": "nome-do-cliente"
  }
  ```

### 2. Chat Completions (OpenAI Compatible)
Endpoint principal para inferência de texto.

- **URL:** `/v1/chat/completions`
- **Método:** `POST`
- **Headers:** `Authorization: Bearer <crom_sk_...>`
- **Body:**
  ```json
  {
    "model": "crom-1",
    "messages": [
      {"role": "user", "content": "Olá, quem é você?"}
    ]
  }
  ```

---

## 🏗️ Estrutura de Segurança
- **API Keys:** Armazenadas como hashes (Argon2id).
- **Master Key:** Utilizada exclusivamente para a administração (criação) de novas chaves de acesso.
- **Worker Pool:** O Go gerencia um pool de instâncias Python para evitar contenção do GIL.

---

## 🛠️ Desenvolvimento
- **Adicionar Dependências:** Use `go get`.
- **Compilar:** `make build`.
- **Testar:** `make test`.
