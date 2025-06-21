package middlewares

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/farovictor/bifrost/config"
	routes "github.com/farovictor/bifrost/routes"
	redis "github.com/redis/go-redis/v9"
)

// Redis connector to manage rate limiting using configuration values.
var rdb = redis.NewClient(&redis.Options{
	Addr:     config.RedisAddr(),
	Password: config.RedisPassword(),
	DB:       config.RedisDB(),
	Protocol: config.RedisProtocol(),
})

type localCounter struct {
	mu   sync.Mutex
	data map[string]struct {
		count int
		ts    time.Time
	}
}

var localRL = &localCounter{data: make(map[string]struct {
	count int
	ts    time.Time
})}

func (l *localCounter) incr(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.data[key]
	if time.Since(e.ts) >= time.Minute {
		e.count = 0
		e.ts = time.Now()
	}
	e.count++
	l.data[key] = e
	return e.count
}

func RateLimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			keyID := r.Header.Get("X-Virtual-Key")
			if keyID == "" {
				keyID = r.URL.Query().Get("key")
			}
			if keyID == "" {
				next.ServeHTTP(w, r)
				return
			}

			vk, err := routes.KeyStore.Get(keyID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			redisKey := fmt.Sprintf("ratelimit:%s:%d", keyID, time.Now().Unix()/60)
			count, err := rdb.Incr(r.Context(), redisKey).Result()
			if err != nil {
				// fallback to local counter when redis is unavailable
				count = int64(localRL.incr(redisKey))
			} else {
				if count == 1 {
					rdb.Expire(r.Context(), redisKey, time.Minute)
				}
			}
			if int(count) > vk.RateLimit {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
