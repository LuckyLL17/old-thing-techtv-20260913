package middleware

import (
	"sync"
	"time"
	"upcycle-hub/config"
	apperr "upcycle-hub/pkg/errors"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	tokens int
	ts     int64
}

type rlStore struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	cfg     *config.RateConfig
}

var rl *rlStore

func RateLimit(cfg *config.RateConfig) gin.HandlerFunc {
	if rl == nil {
		rl = &rlStore{buckets: make(map[string]*bucket), cfg: cfg}
		go rl.cleanup()
	}
	return func(c *gin.Context) {
		if !cfg.Enable {
			c.Next()
			return
		}
		key := c.ClientIP() + ":" + c.Request.URL.Path
		if !rl.allow(key) {
			c.JSON(429, respErr(apperr.ErrRateLimit))
			c.Abort()
			return
		}
		c.Next()
	}
}

func (r *rlStore) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().Unix()
	b, ok := r.buckets[key]
	w := int64(r.cfg.Window)
	if !ok || now-b.ts >= w {
		r.buckets[key] = &bucket{tokens: r.cfg.Limit - 1, ts: now}
		return true
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

func (r *rlStore) cleanup() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for range t.C {
		r.mu.Lock()
		now := time.Now().Unix()
		w := int64(r.cfg.Window) * 2
		for k, b := range r.buckets {
			if now-b.ts > w {
				delete(r.buckets, k)
			}
		}
		r.mu.Unlock()
	}
}
