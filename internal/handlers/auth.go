package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"ccms.com/api/internal/database"
	// server "ccms.com/api/internal/server_v2"
)

type Handler struct {
	db database.Querier
}

func NewHandler(querier database.Querier) *Handler {
	return &Handler{db: querier}
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}
	_, _ = w.Write(jsonResp)
}
