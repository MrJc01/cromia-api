#!/bin/bash
set -e

echo "-> Test: Gerando arquivo de log dos endpoints"
LOG_FILE="tests/endpoints_log.md"

echo "# CromIA API Gateway - Endpoints Log" > $LOG_FILE
echo "Data: $(date)" >> $LOG_FILE
echo "" >> $LOG_FILE

echo "## 1. Rota de Modelos Ativos (/v1/models)" >> $LOG_FILE
echo '```json' >> $LOG_FILE
curl -s -X GET http://localhost:8080/v1/models -H "Authorization: Bearer $API_KEY" >> $LOG_FILE
echo "" >> $LOG_FILE
echo '```' >> $LOG_FILE
echo "" >> $LOG_FILE

echo "## 2. Consulta do Saldo do Usuário (Painel Admin)" >> $LOG_FILE
echo '```json' >> $LOG_FILE
COOKIE=$(cat /tmp/cromia_cookie)
curl -s --cookie "session=$COOKIE" http://localhost:8080/v1/admin/me >> $LOG_FILE
echo "" >> $LOG_FILE
echo "## 3. Rota de Saldo via API (/v1/balance)" >> $LOG_FILE
echo '```json' >> $LOG_FILE
curl -s -H "Authorization: Bearer $API_KEY" http://localhost:8080/v1/balance >> $LOG_FILE
echo "" >> $LOG_FILE
echo '```' >> $LOG_FILE
echo "" >> $LOG_FILE

echo "## 4. Rota de Inferência Proxy (Chat Completions)" >> $LOG_FILE
echo '*(Desabilitado temporariamente nos testes automáticos para não gastar seus $5 reais no servidor real)*' >> $LOG_FILE
# curl -s -X POST http://localhost:8080/v1/chat/completions \
#   -H "Authorization: Bearer $API_KEY" \
#   -H "Content-Type: application/json" \
#   -d '{"model": "deepseek-chat", "messages": [{"role": "user", "content": "Hello"}], "stream": false}' >> $LOG_FILE
echo "" >> $LOG_FILE

echo "## 5. Rota de Estimativa Pré-Voo (/v1/estimate)" >> $LOG_FILE
echo '```json' >> $LOG_FILE
curl -s -X POST http://localhost:8080/v1/estimate \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "deepseek-chat", "messages": [{"role": "user", "content": "Calcule quanto isso vai custar."}], "max_tokens": 500}' >> $LOG_FILE
echo "" >> $LOG_FILE
echo '```' >> $LOG_FILE
echo "" >> $LOG_FILE

echo "## 6. Rota de Histórico de Logs (/v1/usage)" >> $LOG_FILE
echo '```json' >> $LOG_FILE
curl -s -X GET http://localhost:8080/v1/usage \
  -H "Authorization: Bearer $API_KEY" >> $LOG_FILE
echo "" >> $LOG_FILE
echo '```' >> $LOG_FILE
echo "" >> $LOG_FILE

echo "   Passou! Log criado em $LOG_FILE"
