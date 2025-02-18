package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	// Define a simple handler that returns a 200 OK response
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a rate limiter middleware with a rate of 1 request per second and a burst of 1
	rateLimitMiddleware := RateLimitMiddleware(time.Second, 1)
	limitedHandler := rateLimitMiddleware(handler)

	// Create a test server with the limited handler
	server := httptest.NewServer(limitedHandler)
	defer server.Close()

	// Test that the first request is allowed
	req, err := http.NewRequest("GET", server.URL, nil)
	assert.NoError(t, err)
	req.RemoteAddr = "127.0.0.1:1234"
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test that the second request is rate limited
	req, err = http.NewRequest("GET", server.URL, nil)
	assert.NoError(t, err)
	req.RemoteAddr = "127.0.0.1:1234"
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	// Wait for the rate limiter to allow another request
	time.Sleep(time.Second)

	// Test that the third request is allowed after waiting
	req, err = http.NewRequest("GET", server.URL, nil)
	assert.NoError(t, err)
	req.RemoteAddr = "127.0.0.1:1234"
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
