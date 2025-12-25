package ratelimit

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

const maxEntries = 10000

type entry struct {
	count   int
	resetAt time.Time
}

type Limiter struct {
	mu         sync.Mutex
	entries    map[string]*entry
	limit      int
	window     time.Duration
	lastSweep  time.Time
	trustProxy bool
}

func New(limit int, window time.Duration, trustProxy bool) *Limiter {
	return &Limiter{
		entries:    make(map[string]*entry),
		limit:      limit,
		window:     window,
		lastSweep:  time.Now(),
		trustProxy: trustProxy,
	}
}

// Allow checks if request from key should be allowed
func (l *Limiter) Allow(key string) bool {
	// limit <= 0 means disabled
	if l.limit <= 0 {
		return true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// cleanup stale entries periodically or when approaching cap
	if time.Since(l.lastSweep) > l.window*2 || len(l.entries) >= maxEntries {
		for k, e := range l.entries {
			if now.After(e.resetAt) {
				delete(l.entries, k)
			}
		}
		l.lastSweep = now
	}

	e, ok := l.entries[key]
	if !ok || now.After(e.resetAt) {
		// new entry - check cap first
		if len(l.entries) >= maxEntries {
			return false
		}
		l.entries[key] = &entry{
			count:   1,
			resetAt: now.Add(l.window),
		}
		return true
	}

	// existing entry - always allow through rate limit check
	if e.count >= l.limit {
		return false
	}
	e.count++
	return true
}

// Middleware wraps an http.Handler with rate limiting
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := l.clientIP(r)
		if !l.Allow(key) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// WrapFunc wraps an http.HandlerFunc with rate limiting
func (l *Limiter) WrapFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := l.clientIP(r)
		if !l.Allow(key) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

func (l *Limiter) clientIP(r *http.Request) string {
	if l.trustProxy {
		// check X-Forwarded-For (railway, nginx, etc)
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if idx := strings.Index(xff, ","); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}
		// check X-Real-IP
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return xri
		}
	}
	// use remote addr directly
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}
