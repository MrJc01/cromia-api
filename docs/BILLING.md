# Motor de Billing e Faturamento

Este documento explica como o sistema transforma requisições à IA em débito de créditos na conta dos usuários.

## 1. Bloqueio Otimista de Acesso

Quando uma requisição chega:
1. O Middleware autentica a `api_key`.
2. Identifica o `user_id`.
3. Busca o `balance` atual da tabela `users`.
4. Se `balance <= 0`, a requisição é negada instantaneamente com HTTP 402 (Payment Required) e uma mensagem: `"Insufficient credits. Please top up your account."`.

> **Nota:** Não fazemos "reserva" estrita de tokens antes de rodar, pois com APIs de IA não é possível prever exatemente quantos tokens a LLM vai gerar na resposta antes de executá-la.

## 2. Cálculo de Custos

Cada provedor possui preços distintos de Input (Prompt) e Output (Completion). Como padronização de Gateway, a CromIA utiliza a métrica de **"Créditos"** genéricos.

O custo é calculado usando a tabela `provider_models`:
`Total Cost = (Prompt Tokens + Completion Tokens) * Base_Cost_Per_Token * cost_multiplier`

## 3. Extração do "Usage" via Streaming

O maior desafio técnico de faturar LLMs é o streaming (SSE). Quando usamos `stream: true`, os dados vêm em pedaços (chunks). 

Provedores compatíveis com OpenAI (DeepSeek e OpenRouter) enviam um último chunk antes da string `[DONE]`. Este chunk especial inclui a contagem final de tokens, como neste exemplo:
```json
{
  "choices": [],
  "usage": {
    "prompt_tokens": 50,
    "completion_tokens": 120,
    "total_tokens": 170
  }
}
```

O proxy intercepta passivamente a stream byte a byte. Quando identifica a palavra `"usage":`, ele processa o bloco assincronamente e inicia a rotina de dedução na tabela `users` (UPDATE com operação de subtração) e o insert na tabela `usage_logs`.
