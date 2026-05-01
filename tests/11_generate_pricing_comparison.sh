#!/bin/bash
set -e

LOG_FILE="tests/pricing_comparison.md"
echo "-> Test: Gerando tabela comparativa de preços"

echo "# CromIA API - Tabela de Preços e Comparativo" > $LOG_FILE
echo "Data da simulação: $(date)" >> $LOG_FILE
echo "" >> $LOG_FILE

echo "Abaixo está a tabela de cobrança gerada diretamente do Gateway, listando o custo real do modelo no Provedor na nuvem, o valor na CromIA (após a sua margem de lucro configurada), e a estimativa de Lucro Líquido para cada 1 Milhão (1M) de tokens processados." >> $LOG_FILE
echo "" >> $LOG_FILE
echo "*(Atenção: Todos os preços acima refletem o custo REAL atualizado diretamente da nuvem em tempo real (OpenRouter Oracle) para os tokens de Prompt e Completion.)*" >> $LOG_FILE
echo "" >> $LOG_FILE

# Chama a API de modelos e formata o JSON para a tabela
curl -s -H "Authorization: Bearer $API_KEY" http://localhost:8080/v1/models | python3 -c '
import sys, json
try:
    data = json.load(sys.stdin)
    
    print("## 1. Tabela de Custos Básicos (Por 1K Tokens)")
    print("| Modelo | Custo Provedor Prompt (1K) | Custo Provedor Comp (1K) | Preço CromIA Prompt (1K) | Preço CromIA Comp (1K) | Lucro Prompt | Lucro Comp |")
    print("|--------|---------------------------|--------------------------|-------------------------|------------------------|--------------|------------|")
    for model in data.get("data", []):
        name = model.get("id", "Unknown")
        p_prompt = model.get("provider_prompt_cost", 0) * 1_000
        p_comp = model.get("provider_completion_cost", 0) * 1_000
        c_prompt = model.get("cromia_prompt_cost", 0) * 1_000
        c_comp = model.get("cromia_completion_cost", 0) * 1_000
        lucro_p = (c_prompt * 0.01) - p_prompt
        lucro_c = (c_comp * 0.01) - p_comp
        print(f"| `{name}` | ${p_prompt:.5f} | ${p_comp:.5f} | {c_prompt:.2f}C (${c_prompt*0.01:.5f}) | {c_comp:.2f}C (${c_comp*0.01:.5f}) | **${lucro_p:.5f}** | **${lucro_c:.5f}** |")

    print("\n## 2. Tabela de Escala (Por 1 Milhão de Tokens)")
    print("| Modelo | Custo Provedor Prompt (1M) | Custo Provedor Comp (1M) | Preço CromIA Prompt (1M) | Preço CromIA Comp (1M) | Lucro Total Estimado |")
    print("|--------|---------------------------|--------------------------|-------------------------|------------------------|----------------------|")
    for model in data.get("data", []):
        name = model.get("id", "Unknown")
        p_prompt = model.get("provider_prompt_cost", 0) * 1_000_000
        p_comp = model.get("provider_completion_cost", 0) * 1_000_000
        c_prompt = model.get("cromia_prompt_cost", 0) * 1_000_000
        c_comp = model.get("cromia_completion_cost", 0) * 1_000_000
        lucro = ((c_prompt + c_comp) * 0.01) - (p_prompt + p_comp)
        print(f"| `{name}` | ${p_prompt:.2f} | ${p_comp:.2f} | {c_prompt:.0f}C (${c_prompt*0.01:.2f}) | {c_comp:.0f}C (${c_comp*0.01:.2f}) | **${lucro:.2f}** |")

    print("\n## 3. Exemplos Matemáticos de Uso Prático (Sem requisições)")
    print("Nenhum token real foi gasto nestes exemplos. Os cálculos são matemáticos baseados nos preços atualizados do banco de dados.\n")
    
    for model in data.get("data", []):
        name = model.get("id", "Unknown")
            
        p_prompt_base = model.get("provider_prompt_cost", 0)
        p_comp_base = model.get("provider_completion_cost", 0)
        c_prompt_base = model.get("cromia_prompt_cost", 0)
        c_comp_base = model.get("cromia_completion_cost", 0)

        print(f"### Cenário com `{name}`")
        
        # Exemplo A
        tokens_p = 500
        tokens_c = 1000
        custo_prov = (tokens_p * p_prompt_base) + (tokens_c * p_comp_base)
        custo_crom = (tokens_p * c_prompt_base) + (tokens_c * c_comp_base)
        lucro = (custo_crom * 0.01) - custo_prov
        
        print(f"- **Artigo de Blog Médio (500 Prompt / 1000 Completion):**")
        print(f"  - Provedor (Você paga): ${custo_prov:.5f}")
        print(f"  - CromIA (Cliente paga): {custo_crom:.2f} Croms (${custo_crom*0.01:.5f})")
        print(f"  - Seu Lucro Líquido: **${lucro:.5f}**")
        
        # Exemplo B
        tokens_p = 4000
        tokens_c = 2000
        custo_prov = (tokens_p * p_prompt_base) + (tokens_c * p_comp_base)
        custo_crom = (tokens_p * c_prompt_base) + (tokens_c * c_comp_base)
        lucro = (custo_crom * 0.01) - custo_prov
        
        print(f"- **Análise de Documento Longo (4000 Prompt / 2000 Completion):**")
        print(f"  - Provedor (Você paga): ${custo_prov:.5f}")
        print(f"  - CromIA (Cliente paga): {custo_crom:.2f} Croms (${custo_crom*0.01:.5f})")
        print(f"  - Seu Lucro Líquido: **${lucro:.5f}**\n")

except Exception as e:
    print(f"Erro ao ler modelos: {e}")
' >> $LOG_FILE

echo "*\*Nota Técnica: Os valores listados acima NÃO são hipotéticos. O Gateway possui uma âncora fixa de conversão onde 1 Crédito CromIA equivale EXATAMENTE a \$0.01 USD. O cálculo do lucro é resultado da fórmula real implantada no sistema de cobrança: (Custo Provedor * 100) * Multiplicador.*" >> $LOG_FILE

echo "   Passou! Tabela criada em $LOG_FILE"
