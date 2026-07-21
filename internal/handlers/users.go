package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"ccms.com/api/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ProfileResponse defines the user profile data structure
type ProfileResponse struct {
	ID         pgtype.UUID `json:"id"`
	Username   string      `json:"username"`
	Email      string      `json:"email"`
	EmployeeID int32       `json:"employee_id"`
	CreatedAt  time.Time   `json:"created_at"`
}

func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.db.GetUserByID(r.Context(), pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch profile", http.StatusInternalServerError)
		}
		return
	}

	response := ProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProfileRequest defines the payload for profile updates
type UpdateProfileRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
}

// UpdateProfile handles profile updates
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current user data
	currentUser, err := h.db.GetUserByID(r.Context(), pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Prepare update parameters
	updateParams := database.UpdateUserParams{
		ID: pgtype.UUID{Bytes: userID, Valid: true},
	}

	// Only update fields that are provided in the request
	if req.Username != nil {
		// Check if username is available
		if _, err := h.db.GetUserByUsername(r.Context(), *req.Username); err == nil {
			http.Error(w, "Username already taken", http.StatusConflict)
			return
		}
		updateParams.Username = *req.Username
	} else {
		updateParams.Username = currentUser.Username
	}

	if req.Email != nil {
		// Check if email is available
		if _, err := h.db.GetUserByEmail(r.Context(), *req.Email); err == nil {
			http.Error(w, "Email already in use", http.StatusConflict)
			return
		}
		updateParams.Email = *req.Email
	} else {
		updateParams.Email = currentUser.Email
	}

	// Execute update
	updatedUser, err := h.db.UpdateUser(r.Context(), updateParams)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	response := ProfileResponse{
		ID:        updatedUser.ID,
		Username:  updatedUser.Username,
		Email:     updatedUser.Email,
		CreatedAt: updatedUser.CreatedAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
