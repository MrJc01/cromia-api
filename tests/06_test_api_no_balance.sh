#!/bin/bash
set -e
echo "-> Test: Teste de Saldo Insuficiente"
# Remove o saldo do usuário para testar
./cromia users remove-credits --user testuser --amount 600 > /dev/null 2>&1 || true

HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-chat",
    "messages": [{"role": "user", "content": "Hello"}]
  }')

if [ "$HTTP_STATUS" == "402" ]; then
    echo "   Passou! API retornou 402 Payment Required."
else
    echo "   Falhou! API retornou $HTTP_STATUS em vez de 402."
fi
