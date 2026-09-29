package utils

import (
	"sync"
	"time"
)

// RateLimitTier defines different rate limiting levels
type RateLimitTier struct {
	Name        string        `json:"name"`
	Requests    int           `json:"requests"`
	Window      time.Duration `json:"window"`
	Burst       int           `json:"burst"`
	RetryAfter  time.Duration `json:"retry_after"`
	Description string        `json:"description"`
}

// Predefined rate limit tiers
var (
	StrictTier = RateLimitTier{
		Name:        "strict",
		Requests:    5,
		Window:      time.Minute,
		Burst:       3,
		RetryAfter:  60 * time.Second,
		Description: "Auth endpoints rate limit",
	}

	NormalTier = RateLimitTier{
		Name:        "normal",
		Requests:    100,
		Window:      time.Minute,
		Burst:       20,
		RetryAfter:  30 * time.Second,
		Description: "General API rate limit",
	}

	LenientTier = RateLimitTier{
		Name:        "lenient",
		Requests:    1000,
		Window:      time.Minute,
		Burst:       100,
		RetryAfter:  10 * time.Second,
		Description: "Read-heavy endpoints rate limit",
	}

	AdminTier = RateLimitTier{
		Name:        "admin",
		Requests:    500,
		Window:      time.Minute,
		Burst:       50,
		RetryAfter:  15 * time.Second,
		Description: "Admin endpoints rate limit",
	}
)

// RateLimitError represents a structured error response
type RateLimitError struct {
	Error            string `json:"error"`
	Message          string `json:"message"`
	RetryAfter       int    `json:"retry_after_seconds"`
	Limit            int    `json:"rate_limit"`
	Remaining        int    `json:"remaining"`
	Reset            int64  `json:"reset_timestamp"`
	Tier             string `json:"tier,omitempty"`
	DocumentationURL string `json:"documentation_url,omitempty"`
}

// RateLimiter handles rate limiting logic
type RateLimiter struct {
	visitors map[string]*visitorInfo
	mu       sync.RWMutex
	tier     RateLimitTier
}

type visitorInfo struct {
	lastSeen     time.Time
	tokens       float64
	requestCount int64
	userType     string
	mu           sync.Mutex
}

// For Small/Medium Apps:
// Use golang.org/x/time/rate + Redis for distributed limiting

// For Enterprise Apps:
// Consider managed solutions:
// - AWS API Gateway: Built-in rate limiting
// - Cloudflare: Free rate limiting
// - Kong/NGINX: Gateway-level rate limiting
// - Redis Enterprise: Built-in rate limiting modules

// Setup			Idle Memory	Under Load	Req/sec (simple)	Req/sec (complex)
// Go + sqlc		30-80MB		80-200MB	25-40k				10-20k
// Go + GORM		100-200MB	200-400MB	8-15k				5-10k
// Go + sqlx		40-100MB	100-250MB	20-35k				8-18k
// NestJS + Prisma	150-300MB	300-600MB	8-15k				3-8k
// NestJS + TypeORM	200-400MB	400-800MB	5-10k				2-5k
// Next.js + Prisma	250-500MB	500-1000MB	5-10k				2-5k

// Setup				Monthly Cost	Uptime	Complexity	Notes
// Go + sqlc (EC2)		$85-95			99.95%	Low			Best performance/cost
// Go + sqlc (Lambda)	$25-30			99.99%	Medium		Cold starts possible
// Go + GORM (EC2)		$125-150		99.95%	Low			More memory needed
// NestJS (EC2)			$150-170		99.95%	Low			Node.js overhead
// Next.js (EC2)		$165-200		99.95%	Medium		SSR overhead
// Next.js (Vercel)		$25-50			99.99%	Lowest		Easiest deployment

// ---------------------

// Go's stability promise: GUARANTEED backward compatibility
// Go 1.x compatibility promise since 2012 - NEVER broken
// Example: Code written in Go 1.5 (2015) still compiles in Go 1.22 (2024)

// NestJS: BREAKING changes between major versions
// Version 8 → 9 → 10: Each requires code changes

// Next.js: NOTORIOUS for breaking changes
// Major version changes often require significant rewrites
// v15 (2024): React 19, more breaking changes

// ORM (GORM) uses reflection at runtime:
// db.Where("email = ?", email).First(&user)
// Reflection overhead: type checking, struct field mapping
// Memory: stores metadata about all models

// sqlc generates compile-time code:
// user, err := queries.GetUserByEmail(ctx, email)
// No reflection, direct function call
// Memory: only stores query strings
