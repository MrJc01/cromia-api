#!/bin/bash
set -e
echo "-> Test: Web Dashboard Protected Route"
COOKIE=$(cat /tmp/cromia_cookie)
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --cookie "session=$COOKIE" http://localhost:8080/v1/admin/me)

if [ "$HTTP_STATUS" == "200" ]; then
    echo "   Passou! Rota administrativa retornou 200."
else
    echo "   Falhou! HTTP Status: $HTTP_STATUS"
fi
