package providers

import (
	"os"
)

func GetProviderURLAndKey(providerName string) (string, string) {
	if providerName == "deepseek" {
		url := os.Getenv("DEEPSEEK_BASE_URL")
		if url == "" {
			url = "https://api.deepseek.com"
		}
		return url + "/chat/completions", os.Getenv("DEEPSEEK_API_KEY")
	}
	if providerName == "openrouter" {
		url := os.Getenv("OPENROUTER_BASE_URL")
		if url == "" {
			url = "https://openrouter.ai/api/v1"
		}
		return url + "/chat/completions", os.Getenv("OPENROUTER_API_KEY")
	}
	return "", ""
}
