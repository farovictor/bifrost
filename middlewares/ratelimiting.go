package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/FokusInternal/bifrost/config"
	redis "github.com/redis/go-redis/v9"
)

// Redis connector to manage rate limiting using configuration values.
var rdb = redis.NewClient(&redis.Options{
	Addr:     config.RedisAddr(),
	Password: config.RedisPassword(),
	DB:       config.RedisDB(),
	Protocol: config.RedisProtocol(),
})

func RateLimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				http.Error(w, "Missing API key", http.StatusUnauthorized)
				return
			}

			// TODO: This time constraint should be personalized and consider each user quota/plan
			key := fmt.Sprintf("ratelimit:%s:%d", apiKey, time.Now().Unix()/60) // per minute
			count, err := rdb.Incr(r.Context(), key).Result()

			if err != nil {
				http.Error(w, "Redis error", http.StatusInternalServerError)
				return
			}

			if count == 1 {
				rdb.Expire(r.Context(), key, time.Minute)
			}

			// TODO: This counter should be personalized and consider each user quota/plan
			if count > 1000 {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
