package settings

import (
	"net/http"
	"strconv"

	"ccms.com/api/internal/constants"
	"ccms.com/api/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db   database.Querier
	Pool *pgxpool.Pool
}

func NewHandler(db database.Querier, pool *pgxpool.Pool) *Handler {
	return &Handler{
		db:   db,
		Pool: pool,
	}
}

// Shared helpers

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
