# Deploy e Upgrade do CromIA Gateway

A arquitetura do CromIA foi desenhada para facilitar ao máximo a vida do administrador de sistemas, combinando backend, rotas proxy, frontend do dashboard e banco de dados tudo dentro de um único pacote autossuficiente, sem dependências externas (como NPM, Python, PM2 ou Docker).

---

## 1. Como Iniciar o Servidor Pela Primeira Vez

Você não precisa instalar dependências no servidor além do arquivo `.env`.

1. Construa o binário final no seu ambiente de desenvolvimento:
   ```bash
   go build -o cromia api/cmd/server/main.go
   ```
2. Mande o arquivo `cromia`, junto com os scripts `start.sh` e `stop.sh` para o seu servidor VPS.
3. Crie o arquivo `.env` na mesma pasta do binário contendo suas chaves:
   ```env
   PORT=8080
   DB_DRIVER=sqlite3
   DB_DSN=data.db
   DEEPSEEK_API_KEY=sua_chave_aqui
   OPENROUTER_API_KEY=sua_chave_aqui
   MASTER_API_KEY=sua_senha_secreta_aqui
   ```
4. Execute o script de inicialização:
   ```bash
   chmod +x start.sh stop.sh cromia
   ./start.sh
   ```

Isso fará o `nohup` colocar o processo em background e criar o banco de dados `data.db` automaticamente. Os logs de acesso da rede ficarão salvos no arquivo `cromia.log`.

---

## 2. Como Fazer o Upgrade (Atualizar a Versão)

Como não temos banco de dados externos como PostgreSQL para se preocupar e todas as dependências estão compiladas estaticamente, a atualização é literalmente "Substituir e Reiniciar".

**Passo a passo do Upgrade:**

1. No servidor VPS, pare a versão atual rodando o script:
   ```bash
   ./stop.sh
   ```
   *Este script varre o `cromia.pid` e manda um sinal de graceful shutdown para liberar a porta.*

2. Sobrescreva o binário antigo pelo novo binário que você acabou de compilar. 
   ```bash
   cp /caminho/do/download/novo_cromia ./cromia
   chmod +x ./cromia
   ```

3. Suba o servidor novamente:
   ```bash
   ./start.sh
   ```

### E o Banco de Dados?
O arquivo `data.db` ficará intacto no diretório. O novo binário vai se conectar a ele e reconhecer todos os saldos, chaves criptografadas e logs de uso anteriores. Caso você tenha adicionado novas funcionalidades ou tabelas no código-fonte, o ORM inteligente inicial disparará `CREATE TABLE IF NOT EXISTS` para injetar a tabela nova instantaneamente sem causar conflito ou indisponibilidade com os dados legados.
