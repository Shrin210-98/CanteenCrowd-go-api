package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"ccms.com/api/internal/constants"
	"ccms.com/api/internal/database"
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
			if origin == allowed || strings.Contains(origin, "trycloudflare.com") {
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

func RequireAuthMiddleware(queries database.Querier, jwtSecret string) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try Authorization header first
			var token string
			authorizationHeader := r.Header.Get("Authorization")
			if authorizationHeader != "" {
				token = strings.TrimSpace(authorizationHeader)
				if strings.HasPrefix(strings.ToLower(token), "bearer ") {
					parts := strings.SplitN(token, " ", 2)
					token = strings.TrimSpace(parts[1])
				}
			}

			// Fall back to query param (for WebSocket connections)
			if token == "" {
				token = r.URL.Query().Get("token")
			}

			if token == "" {
				utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Authorization required"})
				return
			}

			userID, tenantID, userType, err := utils.ValidateToken(token, jwtSecret)
			if err != nil {
				utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Invalid token"})
				return
			}

			userWithPerms, err := queries.GetUserWithPermissions(r.Context(), database.GetUserWithPermissionsParams{
				ID:       userID,
				TenantID: tenantID,
			})
			if err != nil {
				utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "User not found"})
				return
			}
			// Flatten effective permissions
			var flatPermissions []string
			if userWithPerms.EffectivePermissions != nil {
				var permissions []utils.PermissionNode
				if err := json.Unmarshal(userWithPerms.EffectivePermissions, &permissions); err != nil {
					utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to load permissions"})
					return
				}

				filteredPermissions := utils.FilterEnabledPermissions(permissions)
				flatPermissions = utils.FlattenPermissions(filteredPermissions)
			}

			ctx := context.WithValue(r.Context(), constants.ContextKeyUserID, userID)
			ctx = context.WithValue(ctx, constants.ContextKeyTenantID, tenantID)
			ctx = context.WithValue(ctx, constants.ContextKeyUserType, userType)
			ctx = context.WithValue(ctx, constants.ContextKeyPermissions, flatPermissions)
			// ctx = context.WithValue(ctx, constants.ContextKeyRoleID, userWithPerms.RoleID)
			// ctx = context.WithValue(ctx, constants.ContextKeyHasPermissionOverride, userWithPerms.HasPermissionOverride)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
