package middleware

import (
	"net"
	"net/http"
	"strings"
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

		// Goroutine para agregar tokens al bucket
		go func() {
			ticker := time.NewTicker(rl.rate)
			defer ticker.Stop()
			for range ticker.C {
				rl.mu.Lock()
				if _, exists := rl.visitors[ip]; !exists {
					rl.mu.Unlock()
					return
				}
				select {
				case v.limiter <- time.Now():
				default: // No llenar en exceso
				}
				rl.mu.Unlock()
			}
		}()

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
		ip := getIP(r)
		v := rl.getVisitor(ip)

		select {
		case <-v.limiter:
			next.ServeHTTP(w, r)
		default:
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		}
	})
}

func getIP(r *http.Request) string {
	ip := r.RemoteAddr
	if ip == "" {
		ip = "127.0.0.1"
	}
	if strings.Contains(ip, ":") {
		ip, _, _ = net.SplitHostPort(ip)
	}
	return ip
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rate time.Duration, burst int) func(http.Handler) http.Handler {
	rl := NewRateLimiter(rate, burst)
	return rl.limit
}
