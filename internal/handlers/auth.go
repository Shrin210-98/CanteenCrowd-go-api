package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
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
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid request"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to hash password"})
		return
	}

	user, err := h.db.CreateUser(r.Context(), database.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		UserType:     "owner",
	})

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"data": user, "message": "Failed to Register owner user"})
		return
	}

	responseData := map[string]any{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"lastLogin": user.LastLogin,
		"userType":  user.UserType,
		// "emailVerified": user.EmailVerified,
		// "mfaEnabled":    user.MfaEnabled,
		"createdAt": user.CreatedAt,
		// "updatedAt":     user.UpdatedAt,
	}
	utils.JsonResponse(w, http.StatusCreated, map[string]any{"data": responseData, "message": "Successfully Registered the Owner"})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid Request"})
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), creds.Email)
	if err != nil {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Invalid credentials"})
		return
	}

	if user.IsLocked.Bool && user.LockedUntil.Time.After(time.Now()) {
		utils.JsonResponse(w, http.StatusForbidden, map[string]any{"message": "Account locked"})
		return
	}

	if !utils.CheckPasswordHash(creds.Password, user.PasswordHash) {
		// h.db.IncrementLoginAttempts(r.Context(), user.ID)
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Invalid credentials"})
		return
	}

	// h.db.ResetLoginAttempts(r.Context(), user.ID)

	token, err := utils.GenerateTokens(uuid.UUID(user.ID.Bytes), h.jwtSecret)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to generate token"})
		return
	}

	responseData := map[string]any{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"lastLogin": user.LastLogin,
		"userType":  user.UserType,
		"token":     token,
		// "emailVerified": user.EmailVerified,
		// "mfaEnabled":    user.MfaEnabled,
		"createdAt": user.CreatedAt,
		// "updatedAt":     user.UpdatedAt,
	}
	utils.JsonResponse(w, http.StatusOK, map[string]any{"data": responseData, "message": "Succesfully Logged In"})
}
