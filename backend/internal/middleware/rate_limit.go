package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	l "github.com/seankim658/skullking/internal/logger"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients = make(map[string]*client) // Map to store clients by IP address
	mu      sync.Mutex
)

// Called automatically when the package is imported, runs in a background goroutine
// and periodically cleans up stale entries from the `clients` map to prevent it from
// growing indefinitely.
func init() {
	go func() {
		for {
			time.Sleep(10 * time.Minute) // Interval for cleanup check

			mu.Lock()
			if len(clients) == 0 {
				mu.Unlock()
				continue
			}

			for ip, c := range clients {
				if time.Since(c.lastSeen) > 30*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()
}

// Limits the rate of requests from a single IP address
func RateLimit(rateLimit rate.Limit, burst int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the client's IP address
			ipStr := r.Header.Get("X-Forwarded-For")
			// If not behind a proxy, use RemoteAddr
			if ipStr == "" {
				ipStr, _, _ = net.SplitHostPort(r.RemoteAddr)
			} else {
				ips := strings.Split(ipStr, ",")
				ipStr = strings.TrimSpace(ips[0])
			}

			if ipStr == "" {
				http.Error(w, "Internal Server Error: Could not determine client IP", http.StatusInternalServerError)
				return
			}

			mu.Lock()
			// Check if the client IP exists in the map
			if _, found := clients[ipStr]; !found {
				// Create a new rate limiter for this IP if not found
				clients[ipStr] = &client{limiter: rate.NewLimiter(rateLimit, burst)}
			}
			clients[ipStr].lastSeen = time.Now()
			currentLimiter := clients[ipStr].limiter
			mu.Unlock()

			// Check if the request is allowed by the rate limiter
			if !currentLimiter.Allow() {
				http.Error(w, l.TooManyRequestsError, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
