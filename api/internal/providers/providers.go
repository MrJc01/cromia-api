package providers

import (
	"os"
)

func GetProviderURLAndKey(providerName string) (string, string) {
	if providerName == "deepseek" {
		return "https://api.deepseek.com/chat/completions", os.Getenv("DEEPSEEK_API_KEY")
	}
	if providerName == "openrouter" {
		return "https://openrouter.ai/api/v1/chat/completions", os.Getenv("OPENROUTER_API_KEY")
	}
	return "", ""
}
