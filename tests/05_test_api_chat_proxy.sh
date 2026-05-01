#!/bin/bash
echo "-> Test: API Proxy (Auth e Roteamento)"
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-chat",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": false
  }')

if [ "$HTTP_STATUS" == "200" ]; then
    echo "   Passou! Status obtido do fluxo foi $HTTP_STATUS"
else
    echo "   Falhou! API retornou $HTTP_STATUS"
fi
