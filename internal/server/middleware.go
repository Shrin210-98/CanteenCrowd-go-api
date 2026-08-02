package server

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"ccms.com/api/internal/utils"
)

func RequestLoggerMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	}
}

func CORSMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// log.Printf("CORS Debug - Method: %s, Path: %s, Origin: %s", r.Method, r.URL.Path, origin)

		// Allow your frontend origin
		allowedOrigins := []string{
			"http://localhost:5174",
			"http://localhost:5173",
			"http://127.0.0.1:5174",
			"http://127.0.0.1:5173",
			"http://localhost:3000",
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				break
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func RequireAuthMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			utils.JsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Authorization header required"})
			return
		}

		token := strings.TrimSpace(authorizationHeader)
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			parts := strings.SplitN(token, " ", 2)
			token = strings.TrimSpace(parts[1])
		}

		if token == "" {
			utils.JsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Authorization token required"})
			return
		}

		userID, err := utils.ValidateToken(token, jwtSecret)
		if err != nil {
			utils.JsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Invalid token"})
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type Middleware func(http.Handler) http.HandlerFunc

func MiddlewareChain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next.ServeHTTP
	}
}
