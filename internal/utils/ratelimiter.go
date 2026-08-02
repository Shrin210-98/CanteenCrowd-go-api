package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

// NewRateLimiter creates a new rate limiter with the specified tier
func NewRateLimiter(tier RateLimitTier) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitorInfo),
		tier:     tier,
	}

	// Cleanup old visitors periodically
	go rl.cleanupVisitors()

	return rl
}

// allow checks if a request is allowed and returns rate limit info
func (rl *RateLimiter) allow(ip string) (bool, int, int64) {
	v := rl.getVisitor(ip)
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(v.lastSeen).Seconds()

	// Refill tokens based on time elapsed
	rate := float64(rl.tier.Requests) / rl.tier.Window.Seconds()
	v.tokens = min(float64(rl.tier.Burst), v.tokens+elapsed*rate)

	if v.tokens >= 1 {
		v.tokens--
		v.lastSeen = now
		v.requestCount++

		// Calculate remaining requests and reset time
		remaining := int(v.tokens)
		resetTime := now.Add(rl.tier.Window).Unix()

		return true, remaining, resetTime
	}

	// Calculate retry after time
	resetTime := v.lastSeen.Add(rl.tier.Window).Unix()
	return false, 0, resetTime
}

// getVisitor retrieves or creates a visitor record
func (rl *RateLimiter) getVisitor(ip string) *visitorInfo {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitorInfo{
			tokens:   float64(rl.tier.Burst),
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
	}

	v.lastSeen = time.Now()
	return v
}

// cleanupVisitors removes inactive visitors to prevent memory leaks
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getClientIP extracts the real client IP
func getClientIP(r *http.Request) string {
	// Check for proxy headers first
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Get the first IP in the chain
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	return r.RemoteAddr
}

// getUserType extracts user type from request context or headers
func getUserType(r *http.Request) string {
	// You can customize this based on your auth system
	if userType := r.Header.Get("X-User-Type"); userType != "" {
		return userType
	}

	// Check for API key or token
	if r.Header.Get("Authorization") != "" {
		return "authenticated"
	}

	return "anonymous"
}

// ContextualRateLimiterMiddleware provides rate limiting with context awareness
func ContextualRateLimiterMiddleware(next http.Handler) http.HandlerFunc {
	// Initialize different limiters for different scenarios
	limiters := map[string]*RateLimiter{
		"admin":         NewRateLimiter(AdminTier),
		"premium":       NewRateLimiter(LenientTier),
		"authenticated": NewRateLimiter(NormalTier),
		"anonymous":     NewRateLimiter(StrictTier),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		userType := getUserType(r)

		// Select appropriate limiter based on user type
		limiter := limiters[userType]
		if limiter == nil {
			limiter = limiters["anonymous"]
		}

		// Endpoint-specific adjustments
		tier := limiter.tier
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v1/login") ||
			strings.HasPrefix(r.URL.Path, "/api/v1/register"):
			// Stricter limits for auth endpoints
			tier = StrictTier
		case r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE":
			// Slightly stricter for write operations
			tier.Requests = tier.Requests / 2
		}

		// Check rate limit
		allowed, remaining, resetTime := limiter.allow(ip)

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(tier.Requests))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			// Calculate retry after duration
			retryAfter := int(time.Until(time.Unix(resetTime, 0)).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}

			// Log the rate limit event
			log.Printf("Rate limit exceeded - IP: %s, UserType: %s, Endpoint: %s, Method: %s",
				ip, userType, r.URL.Path, r.Method)

			// Send rich error response
			sendRateLimitError(w, tier, retryAfter, resetTime, r)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// sendRateLimitError sends a structured JSON error response
func sendRateLimitError(w http.ResponseWriter, tier RateLimitTier, retryAfter int, resetTime int64, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	w.Header().Set("X-RateLimit-Retry-After", strconv.Itoa(retryAfter))
	w.WriteHeader(http.StatusTooManyRequests)

	errorResponse := RateLimitError{
		Error:            "rate_limit_exceeded",
		Message:          fmt.Sprintf("Rate limit exceeded. Please try again in %d seconds.", retryAfter),
		RetryAfter:       retryAfter,
		Limit:            tier.Requests,
		Remaining:        0,
		Reset:            resetTime,
		Tier:             tier.Name,
		DocumentationURL: "https://your-api-docs.com/rate-limits",
	}

	// Add user-friendly suggestions
	switch tier.Name {
	case "strict":
		errorResponse.Message = fmt.Sprintf("Too many authentication attempts. Please wait %d seconds before trying again.", retryAfter)
	case "normal":
		errorResponse.Message = fmt.Sprintf("API rate limit reached. Your limit will reset in %d seconds.", retryAfter)
	default:
		errorResponse.Message = fmt.Sprintf("Rate limit exceeded. Retry after %d seconds.", retryAfter)
	}

	json.NewEncoder(w).Encode(errorResponse)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Preset middleware constructors for common use cases
func StrictRateLimiter() func(http.Handler) http.HandlerFunc {
	limiter := NewRateLimiter(StrictTier)
	return createMiddleware(limiter)
}

func NormalRateLimiter() func(http.Handler) http.HandlerFunc {
	limiter := NewRateLimiter(NormalTier)
	return createMiddleware(limiter)
}

func LenientRateLimiter() func(http.Handler) http.HandlerFunc {
	limiter := NewRateLimiter(LenientTier)
	return createMiddleware(limiter)
}

// createMiddleware is a helper to create middleware from a limiter
func createMiddleware(limiter *RateLimiter) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			allowed, remaining, resetTime := limiter.allow(ip)

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.tier.Requests))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

			if !allowed {
				retryAfter := int(time.Until(time.Unix(resetTime, 0)).Seconds())
				if retryAfter < 1 {
					retryAfter = 1
				}
				sendRateLimitError(w, limiter.tier, retryAfter, resetTime, r)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// Main handler setup
func setupHandler(mux *http.ServeMux) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 	// No rate limiting for health checks
		// 	if r.URL.Path == "/api/v1/health" {
		// 		MiddlewareChain(CORSMiddleware, RequestLoggerMiddleware)(mux).ServeHTTP(w, r)
		// 		return
		// 	}

		// 	// Auth endpoints - automatic strict rate limiting (handled in contextual middleware)
		// 	if slices.Contains([]string{"/api/v1/register", "/api/v1/login"}, r.URL.Path) {
		// 		MiddlewareChain(
		// 			CORSMiddleware,
		// 			RequestLoggerMiddleware,
		// 			ContextualRateLimiterMiddleware, // Context-aware rate limiting
		// 		)(mux).ServeHTTP(w, r)
		// 		return
		// 	}

		// 	// Protected Routes with contextual rate limiting and authentication
		// 	MiddlewareChain(
		// 		CORSMiddleware,
		// 		RequestLoggerMiddleware,
		// 		ContextualRateLimiterMiddleware, // Context-aware rate limiting
		// 		RequireAuthMiddleware,
		// 	)(mux).ServeHTTP(w, r)
	})

	return handler
}

// Performance Comparison (requests/second on modest hardware):
// Go:      ~50,000+ req/sec
// NestJS:  ~15,000-25,000 req/sec
// Next.js: ~8,000-15,000 req/sec (API routes)
//          ~25,000-40,000 req/sec (static/SSG pages via CDN)

// Docker Image Sizes:
// Go:      ~15MB (multi-stage build, scratch/alpine)
// NestJS:  ~150MB (Node.js + node_modules)
// Next.js: ~200-350MB (Node.js + Next.js + node_modules + build artifacts)

// Memory Usage (idle):
// Go:      ~10-30MB
// NestJS:  ~50-150MB
// Next.js: ~100-300MB (API + SSR combined)

// Cold Start Time:
// Go:      ~10-50ms
// NestJS:  ~500ms-2s
// Next.js: ~1-5s (especially with SSR)

// Go's goroutines excel at:
// - WebSocket connections (100k+ concurrent)
// - Real-time data processing
// - Background job processing
// - API Gateway/BFF pattern

// ----------------------

// Monthly Infrastructure Cost for 1M requests/day:
// Go API:
//   - 1 t3.micro instance:   $8.50/month
//   - Load balancer:          $16.50/month
//   Total:                   ~$25/month

// NestJS API:
//   - 2 t3.small instances:  $34.00/month
//   - Load balancer:          $16.50/month
//   Total:                   ~$50/month

// Next.js (SSR):
//   - 2 t3.medium instances: $68.00/month
//   - Load balancer:          $16.50/month
//   - Vercel Pro alternative: $20/month (limited)
//   Total:                   ~$85/month or $20/month (Vercel)

// ---------------------

// Go's stability promise: GUARANTEED backward compatibility
// Go 1.x compatibility promise since 2012 - NEVER broken
// Example: Code written in Go 1.5 (2015) still compiles in Go 1.22 (2024)

// NestJS: BREAKING changes between major versions
// Version 8 → 9 → 10: Each requires code changes

// Next.js: NOTORIOUS for breaking changes
// Major version changes often require significant rewrites
// v15 (2024): React 19, more breaking changes
