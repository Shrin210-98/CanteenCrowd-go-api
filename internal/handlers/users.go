package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ProfileResponse defines the user profile data structure
type ProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	TenantID  uuid.UUID `json:"tenant_id"`
	UserType  string    `json:"user_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetUserProfile returns the current user's profile
func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	// Get user from database
	user, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.JsonResponse(w, http.StatusNotFound, map[string]any{"message": "User not found"})
		} else {
			log.Printf("Failed to fetch profile: %v", err)
			utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to fetch profile"})
		}
		return
	}

	// Build response
	response := ProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		TenantID:  user.TenantID,
		UserType:  user.UserType,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	utils.JsonResponse(w, http.StatusOK, map[string]any{
		"data":    response,
		"message": "Profile retrieved successfully",
	})
}

// UpdateProfileRequest defines the payload for profile updates
type UpdateProfileRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
}

// UpdateProfile handles profile updates
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	// Parse request body
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid request body"})
		return
	}

	// Validate at least one field is provided
	if req.Username == nil && req.Email == nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "No fields to update"})
		return
	}

	// Get current user data
	currentUser, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.JsonResponse(w, http.StatusNotFound, map[string]any{"message": "User not found"})
		} else {
			log.Printf("Failed to fetch user: %v", err)
			utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to fetch user"})
		}
		return
	}

	// Prepare update parameters with current values
	updateParams := database.UpdateUserParams{
		ID:       userID,
		Username: currentUser.Username,
		Email:    currentUser.Email,
		UserType: currentUser.UserType,
		TenantID: currentUser.TenantID,
		// Preserve other fields
		LastLogin:          currentUser.LastLogin,
		LoginAttempts:      currentUser.LoginAttempts,
		IsLocked:           currentUser.IsLocked,
		LockedUntil:        currentUser.LockedUntil,
		PasswordResetToken: currentUser.PasswordResetToken,
		TokenExpiry:        currentUser.TokenExpiry,
		MustChangePassword: currentUser.MustChangePassword,
		EmailVerified:      currentUser.EmailVerified,
		MfaEnabled:         currentUser.MfaEnabled,
		LastPasswordChange: currentUser.LastPasswordChange,
	}

	// Update username if provided
	if req.Username != nil {
		// Validate username
		if *req.Username == "" {
			utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Username cannot be empty"})
			return
		}

		// Check if username is available
		exists, err := h.db.CheckUsernameExists(r.Context(), *req.Username)
		if err != nil {
			log.Printf("Failed to check username: %v", err)
			utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to check username availability"})
			return
		}
		if exists && *req.Username != currentUser.Username {
			utils.JsonResponse(w, http.StatusConflict, map[string]any{"message": "Username already taken"})
			return
		}
		updateParams.Username = *req.Username
	}

	// Update email if provided
	if req.Email != nil {
		// Validate email
		if *req.Email == "" {
			utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Email cannot be empty"})
			return
		}

		// Check if email is available
		exists, err := h.db.CheckEmailExists(r.Context(), *req.Email)
		if err != nil {
			log.Printf("Failed to check email: %v", err)
			utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to check email availability"})
			return
		}
		if exists && *req.Email != currentUser.Email {
			utils.JsonResponse(w, http.StatusConflict, map[string]any{"message": "Email already in use"})
			return
		}
		updateParams.Email = *req.Email

		// If email changes, require re-verification
		if *req.Email != currentUser.Email {
			updateParams.EmailVerified = false
		}
	}

	// Execute update
	updatedUser, err := h.db.UpdateUser(r.Context(), updateParams)
	if err != nil {
		log.Printf("Failed to update profile: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to update profile"})
		return
	}

	// Build response
	response := ProfileResponse{
		ID:        updatedUser.ID,
		Username:  updatedUser.Username,
		Email:     updatedUser.Email,
		TenantID:  updatedUser.TenantID,
		UserType:  updatedUser.UserType,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	}

	utils.JsonResponse(w, http.StatusOK, map[string]any{
		"data":    response,
		"message": "Profile updated successfully",
	})
}

// GetUserProfileByID returns a user's profile by ID (admin only)
func (h *Handler) GetUserProfileByID(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL parameter
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid user ID"})
		return
	}

	// Get user from database
	user, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.JsonResponse(w, http.StatusNotFound, map[string]any{"message": "User not found"})
		} else {
			log.Printf("Failed to fetch user: %v", err)
			utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to fetch user"})
		}
		return
	}

	// Build response
	response := ProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		TenantID:  user.TenantID,
		UserType:  user.UserType,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	utils.JsonResponse(w, http.StatusOK, map[string]any{
		"data":    response,
		"message": "Profile retrieved successfully",
	})
}
