package constants

import "time"

// Server configuration
const (
	// API version
	APIVersion = "v1"
	APIPrefix  = "/api/v1"

	// Server timeouts
	ServerReadTimeout  = 10 * time.Second
	ServerWriteTimeout = 30 * time.Second
	ServerIdleTimeout  = 60 * time.Second

	// Database settings
	DBMaxConnections = 20
	DBMinConnections = 5
	DBMaxIdleTime    = 30 * time.Minute
	DBMaxLifetime    = 60 * time.Minute

	// JWT settings
	TokenExpiry = 24 * time.Hour
	TokenType   = "Bearer"
)

// CORS configuration
const (
	CORSMaxAge = "86400"
)

var (
	// AllowedOrigins for CORS
	AllowedOrigins = []string{
		"http://localhost:5174",
		"http://localhost:5173",
		"http://127.0.0.1:5174",
		"http://127.0.0.1:5173",
		"http://localhost:3000",
	}

	// AllowedMethods for CORS
	AllowedMethods = "GET, POST, PUT, DELETE, OPTIONS, PATCH"

	// AllowedHeaders for CORS
	AllowedHeaders = "Content-Type, Authorization"
)

// Context keys
type ContextKey string

const (
	ContextKeyUserID   ContextKey = "userID"
	ContextKeyTenantID ContextKey = "tenantID"
	ContextKeyUserType ContextKey = "userType"
)
