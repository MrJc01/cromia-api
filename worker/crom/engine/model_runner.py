"""
engine/model_runner.py

Motor de inferência do Crom IA.
Tenta usar llama-cpp-python para execução de modelos GGUF reais.
Se não estiver disponível, usa um gerador de texto simulado coerente como fallback.
"""
from __future__ import annotations

import os
import time
from typing import Generator

# ─────────────────────────── Detecção de Backend ─────────────────────────────

_LLAMA_AVAILABLE = False
_llama_model_cache: dict = {}

try:
    from llama_cpp import Llama
    _LLAMA_AVAILABLE = True
except ImportError:
    pass


# ─────────────────────────── Interface Pública ───────────────────────────────

def generate(model: str, messages: list[dict], params: dict, stream: bool = False):
    """
    Gera uma resposta para a lista de mensagens fornecida.

    Args:
        model:    nome do modelo configurado (ex: 'crom-1')
        messages: lista de {'role': ..., 'content': ...}
        params:   parâmetros extras (temperature, max_tokens, etc.)
        stream:   se True, retorna um Generator de dicts de chunk

    Returns:
        dict com {'output', 'usage'} OU Generator[dict] no modo stream
    """
    if _LLAMA_AVAILABLE:
        return _llama_generate(model, messages, params, stream)
    else:
        return _fallback_generate(model, messages, params, stream)


# ─────────────────────────── Backend: llama-cpp ──────────────────────────────

def _get_llama_model(model_name: str) -> "Llama":
    """Carrega e cacheia o modelo GGUF configurado via variável de ambiente."""
    if model_name not in _llama_model_cache:
        model_path = os.environ.get(
            "CROM_MODEL_PATH",
            f"models/{model_name}.gguf"
        )
        if not os.path.exists(model_path):
            raise FileNotFoundError(
                f"Modelo '{model_name}' não encontrado em '{model_path}'. "
                f"Configure CROM_MODEL_PATH ou coloque o .gguf em models/."
            )
        _llama_model_cache[model_name] = Llama(
            model_path=model_path,
            n_ctx=int(os.environ.get("CROM_CTX", "2048")),
            n_threads=int(os.environ.get("CROM_THREADS", "4")),
            verbose=False,
        )
    return _llama_model_cache[model_name]


def _format_prompt(messages: list[dict]) -> str:
    """Formata a lista de mensagens no padrão chat instruct."""
    lines = []
    for msg in messages:
        role = msg.get("role", "user")
        content = msg.get("content", "")
        if role == "system":
            lines.append(f"[SYSTEM] {content}")
        elif role == "user":
            lines.append(f"[USER] {content}")
        elif role == "assistant":
            lines.append(f"[ASSISTANT] {content}")
    lines.append("[ASSISTANT]")
    return "\n".join(lines)


def _llama_generate(model: str, messages: list, params: dict, stream: bool):
    llm = _get_llama_model(model)
    prompt = _format_prompt(messages)

    max_tokens = int(params.get("max_tokens", 512))
    temperature = float(params.get("temperature", 0.7))

    if stream:
        def _stream_gen() -> Generator[dict, None, None]:
            for chunk in llm(
                prompt,
                max_tokens=max_tokens,
                temperature=temperature,
                stream=True,
            ):
                text = chunk["choices"][0].get("text", "")
                if text:
                    yield {"delta": text, "finish_reason": None}
            yield {"delta": "", "finish_reason": "stop"}
        return _stream_gen()

    result = llm(prompt, max_tokens=max_tokens, temperature=temperature)
    text = result["choices"][0]["text"].strip()
    usage = result.get("usage", {})
    return {
        "output": text,
        "usage": {
            "prompt_tokens": usage.get("prompt_tokens", 0),
            "completion_tokens": usage.get("completion_tokens", 0),
        },
    }


# ─────────────────────────── Backend: Fallback ───────────────────────────────

_FALLBACK_RESPONSES = [
    "Olá! Sou o Crom IA. Como posso ajudar você hoje?",
    "Entendido. Processando sua solicitação com base nas informações disponíveis.",
    "Boa pergunta! Com base no contexto fornecido, posso elaborar uma resposta detalhada.",
    "Analisando as mensagens... Aqui está o que posso oferecer com base na conversa.",
    "Sou o Crom IA, em modo de demonstração. Integre um modelo GGUF para respostas reais.",
]


def _fallback_generate(model: str, messages: list, params: dict, stream: bool):
    """Gerador de resposta simulado, mas coerente — usado quando llama_cpp não está instalado."""
    last_user_msg = ""
    for msg in reversed(messages):
        if msg.get("role") == "user":
            last_user_msg = msg.get("content", "")
            break

    # Seleciona resposta baseada no hash do conteúdo para ser determinística
    idx = hash(last_user_msg) % len(_FALLBACK_RESPONSES)
    response_text = (
        f"{_FALLBACK_RESPONSES[idx]} "
        f"(modelo: {model}, mensagens: {len(messages)})"
    )

    prompt_tokens = sum(len(m.get("content", "").split()) for m in messages)
    completion_tokens = len(response_text.split())

    if stream:
        def _stream_gen() -> Generator[dict, None, None]:
            words = response_text.split()
            for i, word in enumerate(words):
                text = word + (" " if i < len(words) - 1 else "")
                yield {"delta": text, "finish_reason": None}
                time.sleep(0.03)  # Simula latência de geração token-a-token
            yield {"delta": "", "finish_reason": "stop"}
        return _stream_gen()

    return {
        "output": response_text,
        "usage": {
            "prompt_tokens": prompt_tokens,
            "completion_tokens": completion_tokens,
        },
    }
