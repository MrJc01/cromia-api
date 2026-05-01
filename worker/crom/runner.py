"""
crom/runner.py

Ponto de entrada para o motor de inferência.
Delega ao engine/model_runner.py.
"""
from __future__ import annotations

from crom.engine import model_runner


def run_model(model: str, messages: list, params: dict = None, stream: bool = False):
    """
    Interface chamada pelo socket_server.py.

    Args:
        model:    nome do modelo (ex: 'crom-1')
        messages: lista de {'role': ..., 'content': ...}
        params:   dict com parâmetros extras (temperature, max_tokens, etc.)
        stream:   se True, retorna um Generator de chunks

    Returns:
        dict {'output', 'usage'} OU Generator[dict] se stream=True
    """
    if params is None:
        params = {}

    return model_runner.generate(model, messages, params, stream=stream)
