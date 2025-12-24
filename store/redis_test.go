package store

import (
	"context"
	"os"
	"testing"
	"time"
)

func getRedisAddr() string {
	if addr := os.Getenv("TEST_REDIS_ADDR"); addr != "" {
		return addr
	}
	return "localhost:16379"
}

func TestRedisStore(t *testing.T) {
	ctx := context.Background()

	r, err := NewRedis(getRedisAddr(), "", "hauk-test")
	if err != nil {
		t.Skipf("redis not available: %v", err)
	}
	defer r.Close()

	t.Run("set and get", func(t *testing.T) {
		data := map[string]string{"foo": "bar"}
		err := r.Set(ctx, "test1", data, time.Now().Add(time.Hour))
		if err != nil {
			t.Fatalf("set failed: %v", err)
		}

		var result map[string]string
		err = r.Get(ctx, "test1", &result)
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if result["foo"] != "bar" {
			t.Errorf("expected bar, got: %s", result["foo"])
		}
	})

	t.Run("get nonexistent", func(t *testing.T) {
		var result string
		err := r.Get(ctx, "nonexistent-redis-key", &result)
		if err == nil {
			t.Error("expected error for nonexistent key")
		}
	})

	t.Run("exists", func(t *testing.T) {
		r.Set(ctx, "exists-test", "value", time.Now().Add(time.Hour))

		exists, err := r.Exists(ctx, "exists-test")
		if err != nil {
			t.Fatalf("exists failed: %v", err)
		}
		if !exists {
			t.Error("expected exists to be true")
		}

		exists, err = r.Exists(ctx, "nonexistent-key")
		if err != nil {
			t.Fatalf("exists failed: %v", err)
		}
		if exists {
			t.Error("expected exists to be false")
		}
	})

	t.Run("delete", func(t *testing.T) {
		r.Set(ctx, "delete-test", "value", time.Now().Add(time.Hour))

		err := r.Delete(ctx, "delete-test")
		if err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		exists, _ := r.Exists(ctx, "delete-test")
		if exists {
			t.Error("expected key to be deleted")
		}
	})

	t.Run("expiration", func(t *testing.T) {
		// set with 1s TTL and poll for expiration
		r.SetTTL(ctx, "ttl-expire-test", "value", time.Second)

		// poll for expiration with timeout
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			exists, _ := r.Exists(ctx, "ttl-expire-test")
			if !exists {
				return // success
			}
			time.Sleep(100 * time.Millisecond)
		}
		t.Error("key did not expire within timeout")
	})

	t.Run("complex types", func(t *testing.T) {
		type nested struct {
			Points [][]float64 `json:"points"`
			Name   string      `json:"name"`
		}
		data := nested{
			Points: [][]float64{{1.5, 2.5}, {3.5, 4.5}},
			Name:   "test",
		}

		err := r.Set(ctx, "complex-redis", data, time.Now().Add(time.Hour))
		if err != nil {
			t.Fatalf("set failed: %v", err)
		}

		var result nested
		err = r.Get(ctx, "complex-redis", &result)
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if result.Name != "test" {
			t.Errorf("expected test, got: %s", result.Name)
		}
		if len(result.Points) != 2 {
			t.Errorf("expected 2 points, got: %d", len(result.Points))
		}
	})

	t.Run("key prefixing", func(t *testing.T) {
		// keys should be prefixed
		r.Set(ctx, "prefix-test", "value", time.Now().Add(time.Hour))

		// the internal key should be "hauk-test-prefix-test"
		exists, _ := r.Exists(ctx, "prefix-test")
		if !exists {
			t.Error("expected prefixed key to exist")
		}
	})
}

func TestRedisIntegration(t *testing.T) {
	ctx := context.Background()

	r, err := NewRedis(getRedisAddr(), "", "hauk-integration")
	if err != nil {
		t.Skipf("redis not available: %v", err)
	}
	defer r.Close()

	// test full session lifecycle like the API does
	type sessionData struct {
		ID       string   `json:"id"`
		Targets  []string `json:"targets"`
		Interval float64  `json:"interval"`
	}

	t.Run("session lifecycle", func(t *testing.T) {
		session := sessionData{
			ID:       "test-session-123",
			Targets:  []string{"share-abc"},
			Interval: 5.0,
		}

		// save session
		err := r.Set(ctx, "session-"+session.ID, session, time.Now().Add(time.Hour))
		if err != nil {
			t.Fatalf("save session failed: %v", err)
		}

		// load session
		var loaded sessionData
		err = r.Get(ctx, "session-"+session.ID, &loaded)
		if err != nil {
			t.Fatalf("load session failed: %v", err)
		}

		if loaded.ID != session.ID {
			t.Errorf("ID mismatch: %s vs %s", loaded.ID, session.ID)
		}
		if len(loaded.Targets) != 1 || loaded.Targets[0] != "share-abc" {
			t.Errorf("targets mismatch: %v", loaded.Targets)
		}

		// update session
		loaded.Targets = append(loaded.Targets, "share-def")
		err = r.Set(ctx, "session-"+loaded.ID, loaded, time.Now().Add(time.Hour))
		if err != nil {
			t.Fatalf("update session failed: %v", err)
		}

		// verify update
		var updated sessionData
		r.Get(ctx, "session-"+loaded.ID, &updated)
		if len(updated.Targets) != 2 {
			t.Errorf("expected 2 targets after update, got: %d", len(updated.Targets))
		}

		// delete session
		err = r.Delete(ctx, "session-"+loaded.ID)
		if err != nil {
			t.Fatalf("delete session failed: %v", err)
		}

		exists, _ := r.Exists(ctx, "session-"+loaded.ID)
		if exists {
			t.Error("session should be deleted")
		}
	})
}
