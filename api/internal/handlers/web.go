package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
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

func (h *WebHandler) authenticateUser(r *http.Request) (int, bool) {
	// 1. Try checking the session cookie
	if cookie, err := r.Cookie("session"); err == nil && cookie.Value != "" {
		if userID, ok := verifyCookie(cookie.Value); ok {
			return userID, true
		}
	}

	// 2. Try checking Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token != "" {
			if userID, ok := verifyCookie(token); ok {
				return userID, true
			}
		}
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

func (h *WebHandler) ServeRegister(w http.ResponseWriter, r *http.Request) {
	content, err := webFS.ReadFile("web/register.html")
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

	match, err := security.ComparePassword(password, u.PasswordHash)
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
	userID, ok := h.authenticateUser(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized"}`))
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
		"is_admin": u.IsAdmin,
	})
}

func (h *WebHandler) APIAdminUsage(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authenticateUser(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized"}`))
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
	userID, ok := h.authenticateUser(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized"}`))
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
	userID, ok := h.authenticateUser(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized"}`))
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
	userID, ok := h.authenticateUser(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized"}`))
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
	userID, ok := h.authenticateUser(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized"}`))
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
	match, err := security.ComparePassword(req.CurrentPassword, u.PasswordHash)
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

func (h *WebHandler) APIRESTLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"Method not allowed"}`))
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Bad request - invalid JSON"}`))
		return
	}

	if req.Username == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Username and password are required"}`))
		return
	}

	u, err := h.DB.GetUserByUsername(req.Username)
	if err != nil || u == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Invalid credentials"}`))
		return
	}

	match, err := security.ComparePassword(req.Password, u.PasswordHash)
	if err != nil || !match {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Invalid credentials"}`))
		return
	}

	token := signCookie(u.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       u.ID,
			"username": u.Username,
			"balance":  u.Balance,
		},
	})
}

func (h *WebHandler) APIRESTRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"Method not allowed"}`))
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Bad request - invalid JSON"}`))
		return
	}

	if len(req.Username) < 3 || len(req.Password) < 6 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Username must be at least 3 characters and password at least 6 characters"}`))
		return
	}

	u, err := h.DB.GetUserByUsername(req.Username)
	if err == nil && u != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error":"Username already taken"}`))
		return
	} else if err != nil && err != sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database error"}`))
		return
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to process password"}`))
		return
	}

	id, err := h.DB.CreateUser(req.Username, hash, 0.0)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to create user"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "User registered successfully with zero balance. To add credits, please contact one of the CROM Guardians (mrj.crom@gmail.com).",
		"user": map[string]interface{}{
			"id":       id,
			"username": req.Username,
			"balance":  0.0,
		},
	})
}

func (h *WebHandler) APIAdminUsersList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"Method not allowed"}`))
		return
	}

	userID, ok := h.authenticateUser(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized"}`))
		return
	}

	u, err := h.DB.GetUserByID(userID)
	if err != nil || u == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized"}`))
		return
	}

	if !u.IsAdmin {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Forbidden - Admin only"}`))
		return
	}

	users, err := h.DB.ListUsers()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve users"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
	})
}

func (h *WebHandler) APIAdminUsersCredits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"Method not allowed"}`))
		return
	}

	userID, ok := h.authenticateUser(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized"}`))
		return
	}

	u, err := h.DB.GetUserByID(userID)
	if err != nil || u == nil || !u.IsAdmin {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Forbidden - Admin only"}`))
		return
	}

	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request payload"}`))
		return
	}

	targetUser, err := h.DB.GetUserByID(req.UserID)
	if err != nil || targetUser == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Target user not found"}`))
		return
	}

	err = h.DB.AddBalance(req.UserID, req.Amount)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to adjust balance"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func (h *WebHandler) APIAdminUsersToggleAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"Method not allowed"}`))
		return
	}

	userID, ok := h.authenticateUser(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized"}`))
		return
	}

	u, err := h.DB.GetUserByID(userID)
	if err != nil || u == nil || !u.IsAdmin {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Forbidden - Admin only"}`))
		return
	}

	var req struct {
		UserID  int  `json:"user_id"`
		IsAdmin bool `json:"is_admin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request payload"}`))
		return
	}

	targetUser, err := h.DB.GetUserByID(req.UserID)
	if err != nil || targetUser == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Target user not found"}`))
		return
	}

	// Prevent self-demotion to ensure at least one admin remains!
	if req.UserID == userID && !req.IsAdmin {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Cannot remove admin privileges from yourself"}`))
		return
	}

	err = h.DB.SetAdminStatus(req.UserID, req.IsAdmin)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to update admin status"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}



