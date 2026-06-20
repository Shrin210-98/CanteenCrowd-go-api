package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"

	"ccms.com/api/internal/database"
	"ccms.com/api/internal/handlers"
)

var (
	dbName    = os.Getenv("DB_DATABASE")
	password  = os.Getenv("DB_PASSWORD")
	username  = os.Getenv("DB_USERNAME")
	port      = os.Getenv("DB_PORT")
	host      = os.Getenv("DB_HOST")
	schema    = os.Getenv("DB_SCHEMA")
	jwtSecret = os.Getenv("JWT_SECRET")
)

func NewServer() *http.Server {

	// -- Connecting Database --
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=%s", host, port, username, password, dbName, schema)
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}
	queries := database.New(conn)

	// -- Routes --
	mux := RegisterRoutes(handlers.NewHandler(queries, jwtSecret, conn))

	// -- Middleware implementation --
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Public Routes
		if slices.Contains([]string{"/api/v1/health", "/api/v1/register", "/api/v1/login"}, r.URL.Path) {
			MiddlewareChain(RequestLoggerMiddleware)(mux).ServeHTTP(w, r)
			return
		}
		// Protected Routes with Authenticaction
		MiddlewareChain(RequestLoggerMiddleware, RequireAuthMiddleware)(mux).ServeHTTP(w, r)
	})

	// -- HTTP Server Instance --
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return server
}
