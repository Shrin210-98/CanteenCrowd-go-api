package requests

import (
	"net/http"
	"strconv"

	"ccms.com/api/internal/constants"
	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db   database.Querier
	Pool *pgxpool.Pool
	hub  *Hub
}

func NewHandler(db database.Querier, pool *pgxpool.Pool) *Handler {
	hub := NewHub()
	go hub.Run() // Start hub in background

	return &Handler{
		db:   db,
		Pool: pool,
		hub:  hub,
	}
}

// Shared helpers
func (h *Handler) getEmployeeByUserID(r *http.Request, userID uuid.UUID) (database.Employee, error) {
	tenantID, _ := getTenantID(r)
	return h.db.GetEmployeeByUserID(r.Context(), database.GetEmployeeByUserIDParams{
		UserID:   &userID,
		TenantID: tenantID,
	})
}

// GetRequestByID returns a single request by ID
func (h *Handler) GetRequestByID(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	requestID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	request, err := h.db.GetRequestByID(r.Context(), database.GetRequestByIDParams{
		ID:       requestID,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetRequestByID", "Request not found")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, request, "Request retrieved successfully")
}

func getTenantID(r *http.Request) (uuid.UUID, bool) {
	tenantID, ok := r.Context().Value(constants.ContextKeyTenantID).(uuid.UUID)
	return tenantID, ok
}

func getPagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
