package middleware

import (
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rate     time.Duration
	burst    int
}

type visitor struct {
	limiter  chan time.Time
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter middleware
func NewRateLimiter(rate time.Duration, burst int) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
	}

	go rl.cleanupVisitors()

	return rl
}

func (rl *rateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := make(chan time.Time, rl.burst)
		for i := 0; i < rl.burst; i++ {
			limiter <- time.Now()
		}
		v = &visitor{limiter: limiter, lastSeen: time.Now()}
		rl.visitors[ip] = v
	}

	v.lastSeen = time.Now()
	return v
}

func (rl *rateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		v := rl.getVisitor(ip)

		select {
		case <-v.limiter:
			next.ServeHTTP(w, r)
			v.limiter <- time.Now().Add(rl.rate)
		default:
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		}
	})
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rate time.Duration, burst int) func(http.Handler) http.Handler {
	rl := NewRateLimiter(rate, burst)
	return rl.limit
}
