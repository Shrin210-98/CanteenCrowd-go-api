package handlers

import (
	"encoding/json"
	"net/http"

	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) ListPositions(w http.ResponseWriter, r *http.Request) {
	positions, err := h.db.ListPositions(r.Context())
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve Positions"})
		return
	}
	if len(positions) == 0 {
		utils.JsonResponse(w, http.StatusOK, map[string]string{"message": "No positions found"})
		return
	}
	utils.JsonResponse(w, http.StatusOK, positions)
}

func (h *Handler) CreatePosition(w http.ResponseWriter, r *http.Request) {
	var positionData database.CreatePositionParams
	if err := json.NewDecoder(r.Body).Decode(&positionData); err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Invalid Request Body"})
		return
	}

	author, err := h.db.CreatePosition(r.Context(), positionData)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create Position"})
		return
	}

	utils.JsonResponse(w, http.StatusCreated,
		map[string]any{"data": author, "message": "Successfully Created Position"})
}

func (h *Handler) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	var positionData database.UpdatePositionParams
	if err := json.NewDecoder(r.Body).Decode(&positionData); err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Invalid Request Body"})
		return
	}

	err := h.db.UpdatePosition(r.Context(), positionData)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Position doesn't exist"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeletePosition(w http.ResponseWriter, r *http.Request) {
	var id pgtype.UUID
	err := id.Scan(r.PathValue("id"))
	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid ID"})
		return
	}

	err = h.db.DeletePosition(r.Context(), id)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Position doesn't exist"})
		return
	}
	utils.JsonResponse(w, http.StatusOK, map[string]string{"message": "Successfully Deleted Position"})
}
