package employees

import (
	"encoding/json"
	"log"
	"net/http"

	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
)

func (h *Handler) ListPositions(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	positions, err := h.db.ListPositions(r.Context(), tenantID)
	if err != nil {
		log.Printf("ListPositions error: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to retrieve Positions"})
		return
	}
	if len(positions) == 0 {
		utils.JsonResponse(w, http.StatusOK, map[string]any{"data": []any{}, "message": "No positions found"})
		return
	}
	utils.JsonResponse(w, http.StatusOK, map[string]any{"data": positions, "message": "Positions retrieved successfully"})
}

func (h *Handler) CreatePosition(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	var req struct {
		Title        string  `json:"title"`
		Description  *string `json:"description,omitempty"`
		Level        *int32  `json:"level,omitempty"`
		IsManagement *bool   `json:"is_management,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("CreatePosition: Invalid request body: %v", err)
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid Request Body"})
		return
	}

	if req.Title == "" {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Position title is required"})
		return
	}

	isManagement := false
	if req.IsManagement != nil {
		isManagement = *req.IsManagement
	}

	position, err := h.db.CreatePosition(r.Context(), database.CreatePositionParams{
		TenantID:     tenantID,
		Title:        req.Title,
		Description:  req.Description,
		Level:        toInt4(req.Level),
		IsManagement: isManagement,
	})
	if err != nil {
		log.Printf("CreatePosition: Database error: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to create Position"})
		return
	}

	utils.JsonResponse(w, http.StatusCreated,
		map[string]any{"data": position, "message": "Successfully Created Position"})
}

func (h *Handler) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid ID"})
		return
	}

	var req struct {
		Title        string  `json:"title"`
		Description  *string `json:"description,omitempty"`
		Level        *int32  `json:"level,omitempty"`
		IsManagement *bool   `json:"is_management,omitempty"`
		IsActive     *bool   `json:"is_active,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("UpdatePosition: Invalid request body: %v", err)
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid Request Body"})
		return
	}

	isManagement := false
	if req.IsManagement != nil {
		isManagement = *req.IsManagement
	}

	position, err := h.db.UpdatePosition(r.Context(), database.UpdatePositionParams{
		ID:           id,
		Title:        req.Title,
		Description:  req.Description,
		Level:        toInt4(req.Level),
		IsManagement: isManagement,
		TenantID:     tenantID,
	})
	if err != nil {
		log.Printf("UpdatePosition error: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Position doesn't exist"})
		return
	}

	utils.JsonResponse(w, http.StatusOK, map[string]any{"data": position, "message": "Successfully Updated Position"})
}

func (h *Handler) DeletePosition(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid ID"})
		return
	}

	err = h.db.DeletePosition(r.Context(), database.DeletePositionParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		log.Printf("DeletePosition error: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Position doesn't exist"})
		return
	}
	utils.JsonResponse(w, http.StatusOK, map[string]any{"message": "Successfully Deleted Position"})
}
