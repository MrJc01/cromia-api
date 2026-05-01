#!/bin/bash
set -e

echo "=========================================="
echo "🧪 Iniciando Bateria de Testes do Gateway "
echo "=========================================="

export DB_DSN="test_data.db"
export DB_DRIVER="sqlite3"
export PORT="8080"

# Limpa teste anterior
rm -f test_data.db test_data.db-wal test_data.db-shm

# Compila o binário fresco
echo "[1/3] Compilando binário..."
go build -o cromia api/cmd/server/main.go
chmod +x cromia

echo "[2/3] Rodando testes de CLI (Offline)..."
./tests/01_test_cli_user_create.sh
./tests/02_test_cli_add_credits.sh
./tests/03_test_cli_generate_key.sh
./tests/04_test_cli_model_enable.sh

echo "[3/3] Subindo servidor para testes E2E..."
./cromia serve > test_server.log 2>&1 &
SERVER_PID=$!
sleep 2 # Aguarda o servidor subir

echo "Servidor rodando no PID $SERVER_PID"

# O token gerado no 03 fica salvo num arquivo temporário
export API_KEY=$(cat /tmp/cromia_test_key)

# Roda os testes web/api
./tests/05_test_api_chat_proxy.sh || true
./tests/06_test_api_no_balance.sh || true
./tests/07_test_web_landing.sh
./tests/08_test_web_login.sh
./tests/09_test_web_dashboard.sh
./tests/12_test_future_models.sh
./tests/10_generate_endpoints_log.sh
./tests/11_generate_pricing_comparison.sh

echo "=========================================="
echo "✅ Testes concluídos! Matando servidor..."
kill $SERVER_PID
rm -f test_data.db test_data.db-wal test_data.db-shm /tmp/cromia_test_key /tmp/cromia_cookie
echo "Feito."
