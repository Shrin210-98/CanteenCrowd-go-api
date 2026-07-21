package handlers

import (
	"context"
	"net/http"

	"ccms.com/api/internal/utils"
)

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (h *Handler) DatabaseHealth(w http.ResponseWriter, r *http.Request) {
	stats := make(map[string]string)

	// Ping the database
	err := h.Conn.Ping(context.Background())
	if err != nil {
		stats["status"] = "down"
		stats["error"] = "db down: " + err.Error()
		utils.JsonResponse(w, http.StatusServiceUnavailable, stats)
		return
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// If using pgxpool, get pool stats
	// if h.Pool != nil {
	//     dbStats := h.Pool.Stat()
	//     stats["acquired_conns"] = strconv.FormatInt(int64(dbStats.AcquiredConns()), 10)
	//     stats["idle_conns"] = strconv.FormatInt(int64(dbStats.IdleConns()), 10)
	//     stats["total_conns"] = strconv.FormatInt(int64(dbStats.TotalConns()), 10)
	//     stats["max_conns"] = strconv.FormatInt(int64(dbStats.MaxConns()), 10)
	//     stats["acquire_count"] = strconv.FormatInt(dbStats.AcquireCount(), 10)
	//     stats["acquire_duration"] = dbStats.AcquireDuration().String()
	//     stats["cancelled_acquire_count"] = strconv.FormatInt(dbStats.CancelledAcquireCount(), 10)
	//     stats["constructing_conns"] = strconv.FormatInt(int64(dbStats.ConstructingConns()), 10)
	//     stats["empty_acquire_count"] = strconv.FormatInt(dbStats.EmptyAcquireCount(), 10)

	//     // Evaluate stats to provide a health message
	//     if dbStats.TotalConns() > 40 {
	//         stats["message"] = "The database is experiencing heavy load."
	//     }
	//     if dbStats.AcquireCount() > 1000 {
	//         stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	//     }
	//     if dbStats.IdleConns() < int32(dbStats.TotalConns())/2 {
	//         stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	//     }
	// }

	utils.JsonResponse(w, http.StatusOK, stats)

}
