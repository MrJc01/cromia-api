# CromIA API - Tabela de Preços e Comparativo
Data da simulação: sex 01 mai 2026 19:03:25 -03

Abaixo está a tabela de cobrança gerada diretamente do Gateway, listando o custo real do modelo no Provedor na nuvem, o valor na CromIA (após a sua margem de lucro configurada), e a estimativa de Lucro Líquido para cada 1 Milhão (1M) de tokens processados.

*(Atenção: Todos os preços acima refletem o custo REAL atualizado diretamente da nuvem em tempo real (OpenRouter Oracle) para os tokens de Prompt e Completion.)*

## 1. Tabela de Custos Básicos (Por 1K Tokens)
| Modelo | Custo Provedor Prompt (1K) | Custo Provedor Comp (1K) | Preço CromIA Prompt (1K) | Preço CromIA Comp (1K) | Lucro Prompt | Lucro Comp |
|--------|---------------------------|--------------------------|-------------------------|------------------------|--------------|------------|
| `deepseek-chat` | $0.00000 | $0.00100 | 0.00C ($0.00000) | 0.15C ($0.00150) | **$0.00000** | **$0.00050** |

## 2. Tabela de Escala (Por 1 Milhão de Tokens)
| Modelo | Custo Provedor Prompt (1M) | Custo Provedor Comp (1M) | Preço CromIA Prompt (1M) | Preço CromIA Comp (1M) | Lucro Total Estimado |
|--------|---------------------------|--------------------------|-------------------------|------------------------|----------------------|
| `deepseek-chat` | $0.00 | $1.00 | 0C ($0.00) | 150C ($1.50) | **$0.50** |

## 3. Exemplos Matemáticos de Uso Prático (Sem requisições)
Nenhum token real foi gasto nestes exemplos. Os cálculos são matemáticos baseados nos preços atualizados do banco de dados.

### Cenário com `deepseek-chat`
- **Artigo de Blog Médio (500 Prompt / 1000 Completion):**
  - Provedor (Você paga): $0.00100
  - CromIA (Cliente paga): 0.15 Croms ($0.00150)
  - Seu Lucro Líquido: **$0.00050**
- **Análise de Documento Longo (4000 Prompt / 2000 Completion):**
  - Provedor (Você paga): $0.00200
  - CromIA (Cliente paga): 0.30 Croms ($0.00300)
  - Seu Lucro Líquido: **$0.00100**

*\*Nota Técnica: Os valores listados acima NÃO são hipotéticos. O Gateway possui uma âncora fixa de conversão onde 1 Crédito CromIA equivale EXATAMENTE a $0.01 USD. O cálculo do lucro é resultado da fórmula real implantada no sistema de cobrança: (Custo Provedor * 100) * Multiplicador.*
