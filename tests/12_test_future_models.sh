#!/bin/bash
set -e
echo "-> Test: Habilitando novos modelos (Gemini 3.1, GPT 5.5, Claude 4.6, DeepSeek v4)"

./cromia models enable --provider openrouter --model google/gemini-3.1-flash-lite --multiplier 2.0 > /dev/null
./cromia models enable --provider openrouter --model google/gemini-3.1-pro --multiplier 2.0 > /dev/null
./cromia models enable --provider openrouter --model openai/gpt-5.5 --multiplier 3.0 > /dev/null
./cromia models enable --provider openrouter --model anthropic/claude-opus-4.6 --multiplier 2.5 > /dev/null
./cromia models enable --provider openrouter --model anthropic/claude-sonnet-4.6 --multiplier 2.0 > /dev/null
./cromia models enable --provider openrouter --model deepseek/deepseek-v4-flash --multiplier 1.5 > /dev/null

echo "   Passou!"
