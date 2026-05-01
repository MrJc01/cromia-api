package handlers

import (
	"cromia/api/internal/db"
	"encoding/json"
	"net/http"
	"time"
)

type ModelsHandler struct {
	DB db.DB
}

type ModelOpenAI struct {
	ID                     string  `json:"id"`
	Object                 string  `json:"object"`
	Created                int64   `json:"created"`
	OwnedBy                string  `json:"owned_by"`
	ProviderPromptCost     float64 `json:"provider_prompt_cost"`
	ProviderCompletionCost float64 `json:"provider_completion_cost"`
	CromiaPromptCost       float64 `json:"cromia_prompt_cost"`
	CromiaCompletionCost   float64 `json:"cromia_completion_cost"`
}

type ModelListResponse struct {
	Object string        `json:"object"`
	Data   []ModelOpenAI `json:"data"`
}

func (h *ModelsHandler) List(w http.ResponseWriter, r *http.Request) {
	activeModels, err := h.DB.GetActiveModels()
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	now := time.Now().Unix()
	var models []ModelOpenAI
	for _, m := range activeModels {
		models = append(models, ModelOpenAI{
			ID:                     m.ModelName,
			Object:                 "model",
			Created:                now,
			OwnedBy:                m.ProviderName,
			ProviderPromptCost:     m.PromptCost,
			ProviderCompletionCost: m.CompletionCost,
			CromiaPromptCost:       m.PromptCost * 100.0 * m.CostMultiplier,
			CromiaCompletionCost:   m.CompletionCost * 100.0 * m.CostMultiplier,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ModelListResponse{
		Object: "list",
		Data:   models,
	})
}
