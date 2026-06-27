package handlers

import (
	"bufio"
	"bytes"
	"cromia/api/internal/db"
	"cromia/api/internal/middleware"
	"cromia/api/internal/providers"
	"cromia/api/internal/utils"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

type ChatHandler struct {
	DB db.DB
}

type chatRequest struct {
	Model    string                   `json:"model"`
	Messages []map[string]interface{} `json:"messages"`
	Stream   bool                     `json:"stream"`
}

type openaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (h *ChatHandler) Completions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		utils.JSONError(w, "payload too large", http.StatusBadRequest)
		return
	}

	var req chatRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		utils.JSONError(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		utils.JSONError(w, "model field required", http.StatusBadRequest)
		return
	}

	activeModels, err := h.DB.GetActiveModels()
	if err != nil {
		utils.JSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var matchedModel *db.ProviderModel
	for i := range activeModels {
		if activeModels[i].ModelName == req.Model {
			matchedModel = &activeModels[i]
			break
		}
	}

	if matchedModel == nil {
		utils.JSONError(w, "model not supported or inactive", http.StatusBadRequest)
		return
	}

	providerURL, providerKey := providers.GetProviderURLAndKey(matchedModel.ProviderName)
	if providerURL == "" || providerKey == "" {
		utils.JSONError(w, "provider not configured properly", http.StatusInternalServerError)
		return
	}

	user := r.Context().Value(middleware.UserContextKey).(*db.User)
	apiKey := r.Context().Value(middleware.APIKeyContextKey).(*db.APIKey)

	isFreeModel := strings.HasSuffix(matchedModel.ModelName, ":free")
	if user.Balance <= 0 && !isFreeModel {
		utils.JSONError(w, "Payment Required: Insufficient balance. Please add credits to use paid models.", http.StatusPaymentRequired)
		return
	}

	proxyReq, err := http.NewRequest("POST", providerURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		utils.JSONError(w, "failed to create upstream request", http.StatusInternalServerError)
		return
	}
	proxyReq.Header.Set("Authorization", "Bearer "+providerKey)
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		utils.JSONError(w, "upstream provider error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if !req.Stream {
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)

		respBodyBytes, _ := io.ReadAll(resp.Body)
		w.Write(respBodyBytes)

		// Parse usage asynchronously ONLY if success
		if resp.StatusCode == 200 {
			go h.extractAndChargeUsage(respBodyBytes, user, apiKey, matchedModel, false)
		}
		return
	}

	// Stream proxy mode
	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.JSONError(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)

	scanner := bufio.NewScanner(resp.Body)
	var finalUsage openaiUsage
	var usageFound bool

	for scanner.Scan() {
		line := scanner.Text()
		
		// Pass to client immediately
		w.Write([]byte(line + "\n"))
		flusher.Flush()

		// Intercept usage string
		if resp.StatusCode == 200 && strings.HasPrefix(line, "data: ") && strings.Contains(line, `"usage":`) && !strings.Contains(line, `[DONE]`) {
			jsonStr := strings.TrimPrefix(line, "data: ")
			var chunk struct {
				Usage *openaiUsage `json:"usage"`
			}
			if err := json.Unmarshal([]byte(jsonStr), &chunk); err == nil && chunk.Usage != nil {
				finalUsage = *chunk.Usage
				usageFound = true
			}
		}
	}

	if usageFound {
		go h.chargeUsage(finalUsage, user, apiKey, matchedModel)
	}
}

func (h *ChatHandler) extractAndChargeUsage(respBody []byte, user *db.User, apiKey *db.APIKey, pm *db.ProviderModel, isStream bool) {
	var resp struct {
		Usage *openaiUsage `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &resp); err == nil && resp.Usage != nil {
		h.chargeUsage(*resp.Usage, user, apiKey, pm)
	}
}

func (h *ChatHandler) chargeUsage(u openaiUsage, user *db.User, apiKey *db.APIKey, pm *db.ProviderModel) {
	if u.TotalTokens == 0 {
		return
	}

	var cost float64
	isFreeModel := strings.HasSuffix(pm.ModelName, ":free")

	if pm.PromptCost == 0 && pm.CompletionCost == 0 {
		// Fallback se o preço não foi sincronizado ainda
		if isFreeModel {
			cost = 0 // Modelo garantidamente gratuito
		} else {
			cost = float64(u.TotalTokens) * 0.0001 * pm.CostMultiplier
		}
	} else {
		// Custo Real em Dólares
		dollarCost := float64(u.PromptTokens)*pm.PromptCost + float64(u.CompletionTokens)*pm.CompletionCost
		// 1 Crom = $0.01 -> Multiplica dólares por 100 para ter a base em Croms
		baseCromCost := dollarCost * 100.0
		// Aplica margem de lucro
		cost = baseCromCost * pm.CostMultiplier
	}

	err := h.DB.DeductBalance(user.ID, cost)
	if err != nil {
		log.Printf("[Billing] Error deducting balance for user %d: %v", user.ID, err)
	}
	h.DB.LogUsage(user.ID, apiKey.ID, pm.ModelName, u.PromptTokens, u.CompletionTokens, cost)
}
