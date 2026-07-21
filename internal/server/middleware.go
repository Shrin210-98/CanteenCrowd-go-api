package server

import (
	"context"
	"log"
	"net/http"
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

func RequireAuthMiddleware(next http.Handler) http.HandlerFunc {
	// return func(w http.ResponseWriter, r *http.Request) {
	// 	// Check if the user is Authenticated
	// 	token := r.Header.Get("Authorization")
	// 	if token != "Bearer token" {
	// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 		return
	// 	}
	// 	next.ServeHTTP(w, r)
	// }
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		userID, err := utils.ValidateToken(token, jwtSecret)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user ID to context
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
