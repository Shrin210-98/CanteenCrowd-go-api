package handlers

import (
	"ccms.com/api/internal/database"
	"ccms.com/api/internal/handlers/employees"
	"ccms.com/api/internal/handlers/requests"
	"ccms.com/api/internal/handlers/settings"
	"ccms.com/api/internal/handlers/users"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db database.Querier
	// Conn      *pgx.Conn
	Pool      *pgxpool.Pool
	jwtSecret string
	Users     *users.Handler
	Employees *employees.Handler
	Requests  *requests.Handler
	Settings  *settings.Handler
}

func NewHandler(querier database.Querier, jwtSecret string, pool *pgxpool.Pool) *Handler {
	return &Handler{
		db:        querier,
		jwtSecret: jwtSecret,
		// Conn:      conn,   conn *pgx.Conn
		Pool:      pool,
		Users:     users.NewHandler(querier, pool),
		Employees: employees.NewHandler(querier, pool),
		Requests:  requests.NewHandler(querier, pool),
		Settings:  settings.NewHandler(querier, pool),
	}
}
