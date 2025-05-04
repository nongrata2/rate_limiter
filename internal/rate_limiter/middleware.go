package rate_limiter

import (
	"net/http"
	"ratelimiter/internal/models"
	"strings"
)

func RateLimitMiddleware(store *BucketStore, defaultLimit models.Limit) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := strings.Cut(r.RemoteAddr, ":")
			bucket := store.GetOrCreate(ip, defaultLimit)

			if !bucket.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
