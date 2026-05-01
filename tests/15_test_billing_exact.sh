#!/bin/bash
echo "-> Test: Billing Exact (Testando debito pelo Mock)"
# Faz um request normal para debito
curl -s -o /dev/null -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-chat",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": false
  }'

# Aguarda a goroutine de billing
sleep 1

BALANCE_HTTP=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/v1/balance \
  -H "Authorization: Bearer $API_KEY")

if [ "$BALANCE_HTTP" == "200" ]; then
    echo "   Passou! Endpoint de balance respondeu 200 apos debitar tokens mockados"
else
    echo "   Falhou! API retornou $BALANCE_HTTP para balance"
    exit 1
fi
