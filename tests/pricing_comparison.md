# CromIA API - Tabela de Preços e Comparativo
Data da simulação: sex 01 mai 2026 18:48:55 -03

Abaixo está a tabela de cobrança gerada diretamente do Gateway, listando o custo real do modelo no Provedor na nuvem, o valor na CromIA (após a sua margem de lucro configurada), e a estimativa de Lucro Líquido para cada 1 Milhão (1M) de tokens processados.

*(Atenção: Todos os preços acima refletem o custo REAL atualizado diretamente da nuvem em tempo real (OpenRouter Oracle) para os tokens de Prompt e Completion.)*

## 1. Tabela de Custos Básicos (Por 1K Tokens)
| Modelo | Custo Provedor Prompt (1K) | Custo Provedor Comp (1K) | Preço CromIA Prompt (1K) | Preço CromIA Comp (1K) | Lucro Prompt | Lucro Comp |
|--------|---------------------------|--------------------------|-------------------------|------------------------|--------------|------------|
| `deepseek-chat` | $0.00032 | $0.00089 | 0.05C ($0.00048) | 0.13C ($0.00133) | **$0.00016** | **$0.00044** |
| `google/gemini-3.1-flash-lite` | $0.00000 | $0.00000 | 0.00C ($0.00000) | 0.00C ($0.00000) | **$0.00000** | **$0.00000** |
| `google/gemini-3.1-pro` | $0.00000 | $0.00000 | 0.00C ($0.00000) | 0.00C ($0.00000) | **$0.00000** | **$0.00000** |
| `openai/gpt-5.5` | $0.00500 | $0.03000 | 1.50C ($0.01500) | 9.00C ($0.09000) | **$0.01000** | **$0.06000** |
| `anthropic/claude-opus-4.6` | $0.00500 | $0.02500 | 1.25C ($0.01250) | 6.25C ($0.06250) | **$0.00750** | **$0.03750** |
| `anthropic/claude-sonnet-4.6` | $0.00300 | $0.01500 | 0.60C ($0.00600) | 3.00C ($0.03000) | **$0.00300** | **$0.01500** |
| `deepseek/deepseek-v4-flash` | $0.00014 | $0.00028 | 0.02C ($0.00021) | 0.04C ($0.00042) | **$0.00007** | **$0.00014** |

## 2. Tabela de Escala (Por 1 Milhão de Tokens)
| Modelo | Custo Provedor Prompt (1M) | Custo Provedor Comp (1M) | Preço CromIA Prompt (1M) | Preço CromIA Comp (1M) | Lucro Total Estimado |
|--------|---------------------------|--------------------------|-------------------------|------------------------|----------------------|
| `deepseek-chat` | $0.32 | $0.89 | 48C ($0.48) | 134C ($1.33) | **$0.60** |
| `google/gemini-3.1-flash-lite` | $0.00 | $0.00 | 0C ($0.00) | 0C ($0.00) | **$0.00** |
| `google/gemini-3.1-pro` | $0.00 | $0.00 | 0C ($0.00) | 0C ($0.00) | **$0.00** |
| `openai/gpt-5.5` | $5.00 | $30.00 | 1500C ($15.00) | 9000C ($90.00) | **$70.00** |
| `anthropic/claude-opus-4.6` | $5.00 | $25.00 | 1250C ($12.50) | 6250C ($62.50) | **$45.00** |
| `anthropic/claude-sonnet-4.6` | $3.00 | $15.00 | 600C ($6.00) | 3000C ($30.00) | **$18.00** |
| `deepseek/deepseek-v4-flash` | $0.14 | $0.28 | 21C ($0.21) | 42C ($0.42) | **$0.21** |

## 3. Exemplos Matemáticos de Uso Prático (Sem requisições)
Nenhum token real foi gasto nestes exemplos. Os cálculos são matemáticos baseados nos preços atualizados do banco de dados.

### Cenário com `deepseek-chat`
- **Artigo de Blog Médio (500 Prompt / 1000 Completion):**
  - Provedor (Você paga): $0.00105
  - CromIA (Cliente paga): 0.16 Croms ($0.00157)
  - Seu Lucro Líquido: **$0.00052**
- **Análise de Documento Longo (4000 Prompt / 2000 Completion):**
  - Provedor (Você paga): $0.00306
  - CromIA (Cliente paga): 0.46 Croms ($0.00459)
  - Seu Lucro Líquido: **$0.00153**

### Cenário com `google/gemini-3.1-flash-lite`
- **Artigo de Blog Médio (500 Prompt / 1000 Completion):**
  - Provedor (Você paga): $0.00000
  - CromIA (Cliente paga): 0.00 Croms ($0.00000)
  - Seu Lucro Líquido: **$0.00000**
- **Análise de Documento Longo (4000 Prompt / 2000 Completion):**
  - Provedor (Você paga): $0.00000
  - CromIA (Cliente paga): 0.00 Croms ($0.00000)
  - Seu Lucro Líquido: **$0.00000**

### Cenário com `google/gemini-3.1-pro`
- **Artigo de Blog Médio (500 Prompt / 1000 Completion):**
  - Provedor (Você paga): $0.00000
  - CromIA (Cliente paga): 0.00 Croms ($0.00000)
  - Seu Lucro Líquido: **$0.00000**
- **Análise de Documento Longo (4000 Prompt / 2000 Completion):**
  - Provedor (Você paga): $0.00000
  - CromIA (Cliente paga): 0.00 Croms ($0.00000)
  - Seu Lucro Líquido: **$0.00000**

### Cenário com `openai/gpt-5.5`
- **Artigo de Blog Médio (500 Prompt / 1000 Completion):**
  - Provedor (Você paga): $0.03250
  - CromIA (Cliente paga): 9.75 Croms ($0.09750)
  - Seu Lucro Líquido: **$0.06500**
- **Análise de Documento Longo (4000 Prompt / 2000 Completion):**
  - Provedor (Você paga): $0.08000
  - CromIA (Cliente paga): 24.00 Croms ($0.24000)
  - Seu Lucro Líquido: **$0.16000**

### Cenário com `anthropic/claude-opus-4.6`
- **Artigo de Blog Médio (500 Prompt / 1000 Completion):**
  - Provedor (Você paga): $0.02750
  - CromIA (Cliente paga): 6.88 Croms ($0.06875)
  - Seu Lucro Líquido: **$0.04125**
- **Análise de Documento Longo (4000 Prompt / 2000 Completion):**
  - Provedor (Você paga): $0.07000
  - CromIA (Cliente paga): 17.50 Croms ($0.17500)
  - Seu Lucro Líquido: **$0.10500**

### Cenário com `anthropic/claude-sonnet-4.6`
- **Artigo de Blog Médio (500 Prompt / 1000 Completion):**
  - Provedor (Você paga): $0.01650
  - CromIA (Cliente paga): 3.30 Croms ($0.03300)
  - Seu Lucro Líquido: **$0.01650**
- **Análise de Documento Longo (4000 Prompt / 2000 Completion):**
  - Provedor (Você paga): $0.04200
  - CromIA (Cliente paga): 8.40 Croms ($0.08400)
  - Seu Lucro Líquido: **$0.04200**

### Cenário com `deepseek/deepseek-v4-flash`
- **Artigo de Blog Médio (500 Prompt / 1000 Completion):**
  - Provedor (Você paga): $0.00035
  - CromIA (Cliente paga): 0.05 Croms ($0.00053)
  - Seu Lucro Líquido: **$0.00018**
- **Análise de Documento Longo (4000 Prompt / 2000 Completion):**
  - Provedor (Você paga): $0.00112
  - CromIA (Cliente paga): 0.17 Croms ($0.00168)
  - Seu Lucro Líquido: **$0.00056**

*\*Nota Técnica: Os valores listados acima NÃO são hipotéticos. O Gateway possui uma âncora fixa de conversão onde 1 Crédito CromIA equivale EXATAMENTE a $0.01 USD. O cálculo do lucro é resultado da fórmula real implantada no sistema de cobrança: (Custo Provedor * 100) * Multiplicador.*
