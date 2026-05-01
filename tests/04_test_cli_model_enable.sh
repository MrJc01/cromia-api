#!/bin/bash
set -e
echo "-> Test: Habilitar Modelos de Provedor"
./cromia models enable --provider deepseek --model deepseek-chat --multiplier 1.5 > /tmp/cli_out
grep "enabled" /tmp/cli_out > /dev/null
echo "   Passou!"
