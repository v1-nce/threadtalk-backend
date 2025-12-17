package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	ips    sync.Map
	limit  rate.Limit
	burst  int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		limit: r,
		burst: b,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		v, ok := rl.ips.Load(ip)

		if !ok {
			limiter := rate.NewLimiter(rl.limit, rl.burst)
			v = &visitor{limiter: limiter, lastSeen: time.Now()}
			rl.ips.Store(ip, v)
		}

		vis := v.(*visitor)
		vis.lastSeen = time.Now()

		if !vis.limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please slow down.",
			})
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(1 * time.Minute)
		rl.ips.Range(func(key, value interface{}) bool {
			v := value.(*visitor)
			if time.Since(v.lastSeen) > 3*time.Minute {
				rl.ips.Delete(key)
			}
			return true
		})
	}
}