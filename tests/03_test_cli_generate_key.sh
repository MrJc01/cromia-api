#!/bin/bash
set -e
echo "-> Test: Gerar API Key"
./cromia keys generate --user testuser --name "Test App" > /tmp/cli_out
grep "API Key generated successfully" /tmp/cli_out > /dev/null

# Extrai a key gerada
KEY=$(grep "Key: " /tmp/cli_out | cut -d ' ' -f 2)
echo $KEY > /tmp/cromia_test_key
echo "   Passou! Key=$KEY"
