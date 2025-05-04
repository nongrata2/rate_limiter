package rate_limiter

import (
	"net/http"
	"ratelimiter/internal/models"
	"ratelimiter/internal/repositories"
	"strings"
)

func RateLimitMiddleware(store *BucketStore, defaultLimit models.Limit, db repositories.DBInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				ip, _, _ := strings.Cut(r.RemoteAddr, ":")
				key = ip
			}

			bucket := store.Get(key)
			if bucket == nil {
				dbClient, err := db.GetClient(r.Context(), key)
				if err == nil {
					bucket = NewTokenBucket(dbClient.Capacity, dbClient.RefillRate, dbClient.Unlimited)
					store.Set(key, bucket)
				} else {
					bucket = store.GetOrCreate(key, defaultLimit)
				}
			}

			if !bucket.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
