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
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type ModelListResponse struct {
	Object string        `json:"object"`
	Data   []ModelOpenAI `json:"data"`
}

func (h *ModelsHandler) List(w http.ResponseWriter, r *http.Request) {
	modelNames, err := h.DB.ListModels()
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	now := time.Now().Unix()
	var models []ModelOpenAI
	for _, name := range modelNames {
		models = append(models, ModelOpenAI{
			ID:      name,
			Object:  "model",
			Created: now,
			OwnedBy: "cromia",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ModelListResponse{
		Object: "list",
		Data:   models,
	})
}
