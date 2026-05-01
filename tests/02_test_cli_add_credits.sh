#!/bin/bash
set -e
echo "-> Test: Adicionar Créditos via CLI"
./cromia users add-credits --user testuser --amount 500 > /tmp/cli_out
grep "Added 500" /tmp/cli_out > /dev/null
echo "   Passou!"
