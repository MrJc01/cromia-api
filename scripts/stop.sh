#!/bin/bash
# stop.sh - Para o CromIA rodando em background

cd "$(dirname "$0")/.."

if [ -f cromia.pid ]; then
    PID=$(cat cromia.pid)
    echo "🛑 Parando o CromIA (PID: $PID)..."
    
    # Envia o sinal de desligamento seguro
    kill $PID
    
    # Remove o arquivo PID
    rm cromia.pid
    echo "✅ Servidor parado com sucesso."
else
    echo "❌ Arquivo cromia.pid não encontrado. O CromIA está rodando?"
    # Tenta achar pela marra caso o arquivo não exista
    PID_MARRA=$(pgrep -f "./cromia serve")
    if [ ! -z "$PID_MARRA" ]; then
        echo "Achei o processo perdido (PID: $PID_MARRA). Matando..."
        kill $PID_MARRA
        echo "✅ Morto."
    fi
fi
