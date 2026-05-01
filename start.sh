#!/bin/bash
# start.sh - Inicia o CromIA em background

if [ -f cromia.pid ]; then
    echo "⚠️ O arquivo cromia.pid já existe. O servidor já está rodando?"
    echo "Para forçar, use ./stop.sh primeiro."
    exit 1
fi

echo "🚀 Iniciando o CromIA Gateway em background..."

# Roda com nohup para que continue rodando mesmo se você fechar o terminal
# Redireciona a saída de logs para cromia.log
nohup ./cromia serve > cromia.log 2>&1 &

# Salva o PID (Process ID) para sabermos quem matar depois
echo $! > cromia.pid

echo "✅ Servidor rodando! (PID: $(cat cromia.pid))"
echo "📄 Logs estão sendo salvos em: cromia.log"
