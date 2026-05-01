#!/bin/bash
set -e
echo "-> Test: Criar Usuário via CLI"
./cromia users create --username testuser --password 123456 --balance 100 > /tmp/cli_out
grep "User created" /tmp/cli_out > /dev/null
echo "   Passou!"
