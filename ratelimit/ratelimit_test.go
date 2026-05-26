package ratelimit

import (
	"net/http"
	"testing"
	"time"
)

func TestClientIP(t *testing.T) {
	req := func(remote, xff, xri string) *http.Request {
		r := &http.Request{RemoteAddr: remote, Header: http.Header{}}
		if xff != "" {
			r.Header.Set("X-Forwarded-For", xff)
		}
		if xri != "" {
			r.Header.Set("X-Real-IP", xri)
		}
		return r
	}

	tests := []struct {
		name       string
		trustProxy bool
		r          *http.Request
		want       string
	}{
		// leftmost entries are attacker-controlled; we must land on the rightmost
		{"spoofed leftmost", true, req("10.0.0.1:1", "1.2.3.4, 9.9.9.9, 203.0.113.7", ""), "203.0.113.7"},
		{"single xff", true, req("10.0.0.1:1", "203.0.113.7", ""), "203.0.113.7"},
		{"xff over xri", true, req("10.0.0.1:1", "203.0.113.7", "8.8.8.8"), "203.0.113.7"},
		{"xri fallback", true, req("10.0.0.1:1", "", "8.8.8.8"), "8.8.8.8"},
		{"no headers", true, req("10.0.0.1:1234", "", ""), "10.0.0.1"},
		// untrusted: headers ignored entirely, even if present
		{"distrust ignores xff", false, req("10.0.0.1:1234", "1.2.3.4", "8.8.8.8"), "10.0.0.1"},
	}

	l := New(10, time.Minute, true)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l.trustProxy = tt.trustProxy
			if got := l.clientIP(tt.r); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAllow(t *testing.T) {
	l := New(2, time.Minute, false)
	if !l.Allow("a") || !l.Allow("a") {
		t.Fatal("first two requests should pass")
	}
	if l.Allow("a") {
		t.Error("third request should be denied")
	}
	if !l.Allow("b") {
		t.Error("a different key should have its own bucket")
	}

	// limit <= 0 disables
	off := New(0, time.Minute, false)
	for i := 0; i < 100; i++ {
		if !off.Allow("x") {
			t.Fatal("disabled limiter should always allow")
		}
	}
}
