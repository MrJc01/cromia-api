package handlers

import (
	"cromia/api/internal/db"
	"encoding/json"
	"net/http"
)

type EstimateHandler struct {
	DB db.DB
}

type estimateRequest struct {
	Model    string                   `json:"model"`
	Messages []map[string]interface{} `json:"messages"`
	MaxTokens *int                    `json:"max_tokens"`
	MaxCompletionTokens *int          `json:"max_completion_tokens"`
}

type estimateResponse struct {
	EstimatedPromptTokens     int     `json:"estimated_prompt_tokens"`
	EstimatedCompletionTokens int     `json:"estimated_completion_tokens"`
	EstimatedTotalTokens      int     `json:"estimated_total_tokens"`
	PromptCostCroms           float64 `json:"prompt_cost_croms"`
	MaxCompletionCostCroms    float64 `json:"max_completion_cost_croms"`
	TotalEstimatedCostCroms   float64 `json:"total_estimated_cost_croms"`
}

func (h *EstimateHandler) EstimateCost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req estimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		http.Error(w, `{"error":"model field required"}`, http.StatusBadRequest)
		return
	}

	activeModels, err := h.DB.GetActiveModels()
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
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
		http.Error(w, `{"error":"model not supported or inactive"}`, http.StatusBadRequest)
		return
	}

	// 1. Estimate Prompt Tokens
	var totalChars int
	for _, msg := range req.Messages {
		if content, ok := msg["content"].(string); ok {
			totalChars += len(content)
		}
	}
	// Heurística rápida: 4 caracteres por token
	estPromptTokens := totalChars / 4
	if estPromptTokens == 0 && totalChars > 0 {
		estPromptTokens = 1
	}

	// 2. Estimate Completion Tokens
	estCompTokens := 500 // fallback
	if req.MaxCompletionTokens != nil {
		estCompTokens = *req.MaxCompletionTokens
	} else if req.MaxTokens != nil {
		estCompTokens = *req.MaxTokens
	}

	// 3. Calculos
	var promptCost, compCost float64

	if matchedModel.PromptCost == 0 && matchedModel.CompletionCost == 0 {
		promptCost = float64(estPromptTokens) * 0.0001 * matchedModel.CostMultiplier
		compCost = float64(estCompTokens) * 0.0001 * matchedModel.CostMultiplier
	} else {
		promptCost = float64(estPromptTokens) * matchedModel.PromptCost * 100.0 * matchedModel.CostMultiplier
		compCost = float64(estCompTokens) * matchedModel.CompletionCost * 100.0 * matchedModel.CostMultiplier
	}

	resp := estimateResponse{
		EstimatedPromptTokens:     estPromptTokens,
		EstimatedCompletionTokens: estCompTokens,
		EstimatedTotalTokens:      estPromptTokens + estCompTokens,
		PromptCostCroms:           promptCost,
		MaxCompletionCostCroms:    compCost,
		TotalEstimatedCostCroms:   promptCost + compCost,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
