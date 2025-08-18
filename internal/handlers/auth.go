package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	// server "ccms.com/api/internal/server_v2"
)

type Handler struct {
	db        database.Querier
	jwtSecret string
}

func NewHandler(querier database.Querier, jwtSecret string) *Handler {
	return &Handler{db: querier, jwtSecret: jwtSecret}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username   string `json:"username"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		EmployeeID int32  `json:"employee_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user, err := h.db.CreateUser(r.Context(), database.CreateUserParams{
		EmployeeID:   req.EmployeeID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	})

	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.db.GetUserByUsername(r.Context(), creds.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if user.IsLocked && user.LockedUntil.Time.After(time.Now()) {
		http.Error(w, "Account locked", http.StatusForbidden)
		return
	}

	if !utils.CheckPassword(creds.Password, user.PasswordHash) {
		h.db.IncrementLoginAttempts(r.Context(), user.ID)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	h.db.ResetLoginAttempts(r.Context(), user.ID)

	token, err := utils.GenerateTokens(user.ID, h.jwtSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := struct {
		Token string               `json:"token"`
		User  database.UserAccount `json:"user"`
	}{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
