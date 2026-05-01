#!/bin/bash
echo "-> Test: API Mock Rate Limit (Simulando 429 do Provedor)"
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-chat",
    "messages": [{"role": "user", "content": "trigger_429"}],
    "stream": false
  }')

if [ "$HTTP_STATUS" == "429" ]; then
    echo "   Passou! Status obtido foi 429"
else
    echo "   Falhou! API retornou $HTTP_STATUS"
    exit 1
fi
