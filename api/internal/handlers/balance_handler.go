package handlers

import (
	"cromia/api/internal/db"
	"cromia/api/internal/middleware"
	"encoding/json"
	"net/http"
)

type BalanceHandler struct {
	DB db.DB
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userInter := r.Context().Value(middleware.UserContextKey)
	if userInter == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	cachedUser := userInter.(*db.User)

	freshUser, err := h.DB.GetUserByID(cachedUser.ID)
	if err != nil || freshUser == nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username": freshUser.Username,
		"balance":  freshUser.Balance,
	})
}
