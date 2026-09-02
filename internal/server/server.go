package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
	sslmode   = os.Getenv("DB_SSLMODE")
	jwtSecret = os.Getenv("JWT_SECRET")
)

func NewServer() *http.Server {
	if schema == "" {
		schema = "public"
	}

	// -- Connecting Database --
	// connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=%s", host, port, username, password, dbName, schema)
	// conn, err := pgx.Connect(context.Background(), connStr)
	// if err != nil {
	// 	log.Printf("database connection failed during startup: %v", err)
	// 	conn = nil
	// } else {
	// 	log.Printf("database connection established successfully")
	// }
	// queries := database.New(conn)
	// // -- Routes --
	// mux := RegisterRoutes(handlers.NewHandler(queries, jwtSecret, conn))

	// -- Connecting Database Pool --
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s", host, port, username, password, dbName, sslmode, schema)

	// Create connection pool
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Printf("database connection failed during startup: %v", err)
		pool = nil
	} else {
		// Test the connection
		pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := pool.Ping(pingCtx); err != nil {
			log.Printf("database ping failed: %v", err)
			pool.Close()
			pool = nil
		} else {
			log.Printf("database connection pool established successfully")
		}
	}

	queries := database.New(pool)

	// -- Routes --
	mux := RegisterRoutes(handlers.NewHandler(queries, jwtSecret, pool))

	// -- Middleware implementation --
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Public Routes
		if slices.Contains([]string{"/api/v1/health", "/api/v1/register", "/api/v1/login"}, r.URL.Path) {
			MiddlewareChain(CORSMiddleware, RequestLoggerMiddleware)(mux).ServeHTTP(w, r)
			return
		}
		// Protected Routes with Authenticaction
		MiddlewareChain(CORSMiddleware, RequestLoggerMiddleware, RequireAuthMiddleware)(mux).ServeHTTP(w, r)
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
