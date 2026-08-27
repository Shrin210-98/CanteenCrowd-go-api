package server

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"ccms.com/api/internal/constants"
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

		for _, allowed := range constants.AllowedOrigins {
			if origin == allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				break
			}
		}

		// Use constants for CORS headers
		w.Header().Set("Access-Control-Allow-Methods", constants.AllowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", constants.AllowedHeaders)
		w.Header().Set("Access-Control-Max-Age", constants.CORSMaxAge)

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
			utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Authorization header required"})
			return
		}

		token := strings.TrimSpace(authorizationHeader)
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			parts := strings.SplitN(token, " ", 2)
			token = strings.TrimSpace(parts[1])
		}

		if token == "" {
			utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Authorization token required"})
			return
		}

		userID, tenantID, err := utils.ValidateToken(token, jwtSecret)
		if err != nil {
			utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Invalid token"})
			return
		}

		ctx := context.WithValue(r.Context(), constants.ContextKeyUserID, userID)
		ctx = context.WithValue(ctx, constants.ContextKeyTenantID, tenantID)
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
