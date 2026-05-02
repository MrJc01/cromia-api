package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"cromia/api/internal/db"
	"cromia/api/internal/security"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"embed"
)

//go:embed web/*
var webFS embed.FS

type WebHandler struct {
	DB db.DB
}

func signCookie(userID int) string {
	secret := os.Getenv("MASTER_API_KEY") // use master key as secret
	if secret == "" {
		secret = "default-secret-change-me"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d", userID)))
	return fmt.Sprintf("%d.%s", userID, hex.EncodeToString(mac.Sum(nil)))
}

func verifyCookie(cookieValue string) (int, bool) {
	parts := strings.Split(cookieValue, ".")
	if len(parts) != 2 {
		return 0, false
	}
	userID, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	expected := signCookie(userID)
	if cookieValue == expected {
		return userID, true
	}
	return 0, false
}

func (h *WebHandler) ServeHome(w http.ResponseWriter, r *http.Request) {
	content, err := webFS.ReadFile("web/index.html")
	if err != nil {
		http.Error(w, "Not found", 404)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

func (h *WebHandler) ServeLogin(w http.ResponseWriter, r *http.Request) {
	content, err := webFS.ReadFile("web/login.html")
	if err != nil {
		http.Error(w, "Not found", 404)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

func (h *WebHandler) ServeDocs(w http.ResponseWriter, r *http.Request) {
	content, err := webFS.ReadFile("web/docs.html")
	if err != nil {
		http.Error(w, "Not found", 404)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

func (h *WebHandler) ServeDashboard(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if _, ok := verifyCookie(cookie.Value); !ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	content, err := webFS.ReadFile("web/dashboard.html")
	if err != nil {
		http.Error(w, "Not found", 404)
		return
	}
	
	// Simply serve the html, frontend will fetch data via API
	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

func (h *WebHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	u, err := h.DB.GetUserByUsername(username)
	if err != nil || u == nil {
		http.Error(w, "Invalid credentials", 401)
		return
	}

	match, err := security.CompareAPIKey(password, u.PasswordHash)
	if err != nil || !match {
		http.Error(w, "Invalid credentials", 401)
		return
	}
	
	cookie := http.Cookie{
		Name:     "session",
		Value:    signCookie(u.ID),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func (h *WebHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *WebHandler) APIAdminMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized", 401)
		return
	}
	userID, ok := verifyCookie(cookie.Value)
	if !ok {
		http.Error(w, "Unauthorized", 401)
		return
	}

	u, err := h.DB.GetUserByID(userID)
	if err != nil || u == nil {
		http.Error(w, "User not found", 404)
		return
	}

	// Ideally we also fetch user API keys
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       u.ID,
		"username": u.Username,
		"balance":  u.Balance,
	})
}

func (h *WebHandler) APIAdminUsage(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized", 401)
		return
	}
	userID, ok := verifyCookie(cookie.Value)
	if !ok {
		http.Error(w, "Unauthorized", 401)
		return
	}

	logs, err := h.DB.GetUserUsageLogs(userID)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if logs == nil {
		logs = []db.UsageLog{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": logs,
	})
}

func (h *WebHandler) APIAdminKeysList(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized", 401)
		return
	}
	userID, ok := verifyCookie(cookie.Value)
	if !ok {
		http.Error(w, "Unauthorized", 401)
		return
	}

	keys, err := h.DB.GetUserKeys(userID)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if keys == nil {
		keys = []db.APIKey{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": keys,
	})
}

func (h *WebHandler) APIAdminKeysCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized", 401)
		return
	}
	userID, ok := verifyCookie(cookie.Value)
	if !ok {
		http.Error(w, "Unauthorized", 401)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Name == "" {
		req.Name = "Nova Chave"
	}

	keyString, keyHash, err := security.GenerateAPIKey()
	if err != nil {
		http.Error(w, "Failed to generate key", 500)
		return
	}

	id, err := h.DB.CreateKey(userID, req.Name, keyHash)
	if err != nil {
		http.Error(w, "Failed to save key", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"name":       req.Name,
		"key_string": keyString,
	})
}

func (h *WebHandler) APIAdminKeysRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", 405)
		return
	}
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized", 401)
		return
	}
	userID, ok := verifyCookie(cookie.Value)
	if !ok {
		http.Error(w, "Unauthorized", 401)
		return
	}

	// We extract ID from URL /v1/admin/keys/123
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Bad request", 400)
		return
	}
	keyIDStr := parts[4]
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		http.Error(w, "Bad request", 400)
		return
	}

	// Ensure the key belongs to the user before revoking!
	keys, err := h.DB.GetUserKeys(userID)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	found := false
	for _, k := range keys {
		if k.ID == keyID {
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Forbidden or key not found", 403)
		return
	}

	err = h.DB.RevokeKey(keyID)
	if err != nil {
		http.Error(w, "Failed to revoke key", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebHandler) APIAdminPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized", 401)
		return
	}
	userID, ok := verifyCookie(cookie.Value)
	if !ok {
		http.Error(w, "Unauthorized", 401)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", 400)
		return
	}

	u, err := h.DB.GetUserByID(userID)
	if err != nil || u == nil {
		http.Error(w, "User not found", 404)
		return
	}

	// Verifica se a senha atual confere
	match, err := security.CompareAPIKey(req.CurrentPassword, u.PasswordHash)
	if err != nil || !match {
		http.Error(w, "Invalid current password", 403)
		return
	}

	newHash, err := security.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "Failed to hash password", 500)
		return
	}

	err = h.DB.UpdatePassword(userID, newHash)
	if err != nil {
		http.Error(w, "Failed to update password", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
}
