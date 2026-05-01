package handlers

import (
	"crypto/subtle"
	"cromia/api/internal/db"
	"cromia/api/internal/security"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type KeyHandler struct {
	DB db.DB
}

type CreateKeyRequest struct {
	Name string `json:"name"`
}

func (h *KeyHandler) checkAdminAuth(w http.ResponseWriter, r *http.Request) bool {
	masterKey := []byte("Bearer " + os.Getenv("MASTER_API_KEY"))
	authHeader := []byte(r.Header.Get("Authorization"))

	if subtle.ConstantTimeCompare(authHeader, masterKey) != 1 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

func (h *KeyHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdminAuth(w, r) {
		return
	}

	var req CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	newKey := "crom_sk_" + security.GenerateRandomString(24)
	keyHash, _ := security.HashAPIKey(newKey)

	if err := h.DB.SaveKey(req.Name, keyHash); err != nil {
		http.Error(w, "Failed to save key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"key": newKey})
}

func (h *KeyHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdminAuth(w, r) {
		return
	}

	keys, err := h.DB.GetActiveKeys()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Remove hashes from response for security
	type keyResponse struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
	}
	var resp []keyResponse
	for _, k := range keys {
		resp = append(resp, keyResponse{
			ID:        k.ID,
			Name:      k.Name,
			CreatedAt: k.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *KeyHandler) RevokeKey(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdminAuth(w, r) {
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		// Fallback para ServeMux antigo se não houver path value
		parts := strings.Split(r.URL.Path, "/")
		idStr = parts[len(parts)-1]
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Bad request: invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.DB.RevokeKey(id); err != nil {
		http.Error(w, "Failed to revoke key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
