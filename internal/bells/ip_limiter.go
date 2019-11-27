package bells

import (
	"net/http"
	"sync"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

// IPRateLimiter represents a collection of rate limiters for IP addresses
type IPRateLimiter struct {
	ips    map[string]*rate.Limiter // Normally you'd want this to live in Redis
	mutex  *sync.RWMutex
	limit  rate.Limit
	bucket int
}

// NewIPRateLimiter creates a new collection of rate limiters for IP addresses
func NewIPRateLimiter(limit rate.Limit, bucket int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		ips:    make(map[string]*rate.Limiter),
		mutex:  &sync.RWMutex{},
		limit:  limit,
		bucket: bucket,
	}

	return limiter
}

// AddIP creates a new rate limiter for an IP, and adds it to the collection
func (limiter *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	IPLimiter := rate.NewLimiter(limiter.limit, limiter.bucket)

	limiter.ips[ip] = IPLimiter
	return IPLimiter
}

// GetLimiter fetches a rate limiter for a given IP from the collection
// and otherwise adds one to the collection if not found
func (limiter *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	limiter.mutex.Lock()

	IPLimiter, exists := limiter.ips[ip]

	if !exists {
		limiter.mutex.Unlock()
		return limiter.AddIP(ip)
	}

	limiter.mutex.Unlock()
	return IPLimiter
}

// RateLimitIP is some middleware to rate limit requests by IP address.
func RateLimitIP(h http.Handler) http.Handler {
	limiter := NewIPRateLimiter(0.016, 5) // Approximately recovering 1 token per minute
	fn := func(w http.ResponseWriter, r *http.Request) {
		l := limiter.GetLimiter(r.RemoteAddr)
		if !l.Allow() {
			w.WriteHeader(http.StatusUnauthorized) // Use 401 over 429 due to rate limiting discovery
			LogEntrySetField(r, "failure_reason", "rate_limited_ip")
			// To avoid leaking information, lets do some useless work.
			bcrypt.GenerateFromPassword([]byte("much-wow-such-work"), 8)
			return
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
