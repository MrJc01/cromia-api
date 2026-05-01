package handlers

import (
	"cromia/api/internal/db"
	"cromia/api/internal/middleware"
	"encoding/json"
	"net/http"
)

type UsageHandler struct {
	DB db.DB
}

func (h *UsageHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	user := r.Context().Value(middleware.UserContextKey).(*db.User)

	logs, err := h.DB.GetUserUsageLogs(user.ID)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if logs == nil {
		// return empty array instead of null
		logs = []db.UsageLog{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": logs,
	})
}
