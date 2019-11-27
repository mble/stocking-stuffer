package bells

import (
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

// DeviceRateLimiter represents a collection of rate limiters for devices
type DeviceRateLimiter struct {
	devices map[string]*rate.Limiter // Normally you'd want this to live in Redis
	mutex   *sync.RWMutex
	limit   rate.Limit
	bucket  int
}

// NewDeviceRateLimiter creates a new collection of rate limiters for devices
func NewDeviceRateLimiter(limit rate.Limit, bucket int) *DeviceRateLimiter {
	limiter := &DeviceRateLimiter{
		devices: make(map[string]*rate.Limiter),
		mutex:   &sync.RWMutex{},
		limit:   limit,
		bucket:  bucket,
	}

	return limiter
}

// AddDevice creates a new rate limiter for a device, and adds it to the collection
func (limiter *DeviceRateLimiter) AddDevice(device string) *rate.Limiter {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	DeviceLimiter := rate.NewLimiter(limiter.limit, limiter.bucket)

	limiter.devices[device] = DeviceLimiter
	return DeviceLimiter
}

// GetLimiter fetches a rate limiter for a given device from the collection
// and otherwise adds one to the collection if not found
func (limiter *DeviceRateLimiter) GetLimiter(device string) *rate.Limiter {
	limiter.mutex.Lock()

	DeviceLimiter, exists := limiter.devices[device]

	if !exists {
		limiter.mutex.Unlock()
		return limiter.AddDevice(device)
	}

	limiter.mutex.Unlock()
	return DeviceLimiter
}

// RateLimitDevice is some middleware to rate limit requests by devices.
func RateLimitDevice(h http.Handler) http.Handler {
	limiter := NewDeviceRateLimiter(0.016, 5) // Approximately recovering 1 token per minute
	fn := func(w http.ResponseWriter, r *http.Request) {
		device := r.Header.Get("User-Agent") + "-" + r.Header.Get("Accept-Language") + "-" + r.Header.Get("Accept-Encoding")
		deviceHash := fmt.Sprintf("%x", sha256sum(device))
		LogEntrySetField(r, "device_hash", deviceHash)

		l := limiter.GetLimiter(deviceHash)
		if !l.Allow() {
			w.WriteHeader(http.StatusUnauthorized) // Use 401 over 429 due to rate limiting discovery
			LogEntrySetField(r, "failure_reason", "rate_limited_device")
			// To avoid leaking information, lets do some useless work.
			bcrypt.GenerateFromPassword([]byte("much-wow-such-work"), 8)
			return
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
