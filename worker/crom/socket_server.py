"""
crom/socket_server.py

Servidor Unix Socket que recebe jobs do Go e executa inferência via runner.py.
Suporta dois modos:
  - Normal:    Envia um único JSON com a resposta completa.
  - Streaming: Envia múltiplos JSON line-delimited (um por chunk).
"""
import socket
import json
import os
import sys

import crom.runner


def handle_request(conn: socket.socket, request: dict) -> None:
    """Processa uma requisição e envia a resposta pelo socket."""
    model = request.get("model", "crom-1")
    messages = request.get("messages", [])
    params = request.get("params") or {}
    stream = bool(request.get("stream", False))

    result = crom.runner.run_model(model, messages, params, stream=stream)

    if stream:
        # Modo streaming: envia um JSON por linha até o fim
        try:
            for chunk in result:
                line = json.dumps(chunk, ensure_ascii=False)
                conn.sendall((line + "\n").encode("utf-8"))
        except (BrokenPipeError, ConnectionResetError):
            pass  # Cliente desconectou antes do fim
    else:
        # Modo normal: envia um único JSON
        response_str = json.dumps(result, ensure_ascii=False)
        conn.sendall(response_str.encode("utf-8"))


def start_server(socket_path: str) -> None:
    if os.path.exists(socket_path):
        os.remove(socket_path)

    server = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    server.bind(socket_path)
    server.listen(128)

    # Sinaliza para o Go que o servidor está pronto
    sys.stdout.write("READY\n")
    sys.stdout.flush()

    while True:
        conn, _ = server.accept()
        try:
            # Lê a requisição completa (até 1MB)
            data = b""
            while True:
                chunk = conn.recv(65536)
                if not chunk:
                    break
                data += chunk
                # Tenta parsear: se o JSON estiver completo, para de ler
                try:
                    json.loads(data.decode("utf-8"))
                    break
                except json.JSONDecodeError:
                    continue

            if not data:
                continue

            request = json.loads(data.decode("utf-8"))
            handle_request(conn, request)

        except Exception as e:
            try:
                conn.sendall(json.dumps({"error": str(e)}).encode("utf-8"))
            except Exception:
                pass
        finally:
            conn.close()


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Uso: socket_server.py <socket_path>", file=sys.stderr)
        sys.exit(1)

    socket_path = sys.argv[1]
    start_server(socket_path)
