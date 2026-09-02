package handlers

import (
	"ccms.com/api/internal/database"
	"ccms.com/api/internal/handlers/employees"
	"ccms.com/api/internal/handlers/users"
	"github.com/jackc/pgx/v5/pgxpool"
	// server "ccms.com/api/internal/server_v2"
)

type Handler struct {
	db database.Querier
	// Conn      *pgx.Conn
	Pool      *pgxpool.Pool
	jwtSecret string
	Employees *employees.Handler
	Users     *users.Handler
}

func NewHandler(querier database.Querier, jwtSecret string, pool *pgxpool.Pool) *Handler {
	return &Handler{
		db:        querier,
		jwtSecret: jwtSecret,
		// Conn:      conn,   conn *pgx.Conn
		Pool:      pool,
		Employees: employees.NewHandler(querier, pool),
		Users:     users.NewHandler(querier, pool),
	}
}
