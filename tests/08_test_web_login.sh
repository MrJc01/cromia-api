#!/bin/bash
set -e
echo "-> Test: Web Login Session"
curl -s -i -X POST -d "username=testuser&password=123456" http://localhost:8080/login > /tmp/login_out
grep "Set-Cookie: session=" /tmp/login_out > /dev/null

# Extrai cookie
COOKIE=$(grep "Set-Cookie: session=" /tmp/login_out | cut -d '=' -f 2 | cut -d ';' -f 1)
echo $COOKIE > /tmp/cromia_cookie
echo "   Passou! Cookie interceptado."
