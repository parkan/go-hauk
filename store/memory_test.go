package store

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()

	t.Run("set and get", func(t *testing.T) {
		data := map[string]string{"foo": "bar"}
		err := m.Set(ctx, "test1", data, time.Now().Add(time.Hour))
		if err != nil {
			t.Fatalf("set failed: %v", err)
		}

		var result map[string]string
		err = m.Get(ctx, "test1", &result)
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if result["foo"] != "bar" {
			t.Errorf("expected bar, got: %s", result["foo"])
		}
	})

	t.Run("get nonexistent", func(t *testing.T) {
		var result string
		err := m.Get(ctx, "nonexistent", &result)
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("exists", func(t *testing.T) {
		m.Set(ctx, "exists-test", "value", time.Now().Add(time.Hour))

		exists, err := m.Exists(ctx, "exists-test")
		if err != nil {
			t.Fatalf("exists failed: %v", err)
		}
		if !exists {
			t.Error("expected exists to be true")
		}

		exists, err = m.Exists(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("exists failed: %v", err)
		}
		if exists {
			t.Error("expected exists to be false")
		}
	})

	t.Run("delete", func(t *testing.T) {
		m.Set(ctx, "delete-test", "value", time.Now().Add(time.Hour))

		err := m.Delete(ctx, "delete-test")
		if err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		exists, _ := m.Exists(ctx, "delete-test")
		if exists {
			t.Error("expected key to be deleted")
		}
	})

	t.Run("expiration", func(t *testing.T) {
		m.Set(ctx, "expire-test", "value", time.Now().Add(-time.Second))

		var result string
		err := m.Get(ctx, "expire-test", &result)
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound for expired key, got: %v", err)
		}

		exists, _ := m.Exists(ctx, "expire-test")
		if exists {
			t.Error("expected expired key to not exist")
		}
	})

	t.Run("set with ttl", func(t *testing.T) {
		err := m.SetTTL(ctx, "ttl-test", "value", time.Hour)
		if err != nil {
			t.Fatalf("setTTL failed: %v", err)
		}

		exists, _ := m.Exists(ctx, "ttl-test")
		if !exists {
			t.Error("expected key to exist")
		}
	})

	t.Run("clear", func(t *testing.T) {
		m.Set(ctx, "clear-test", "value", time.Now().Add(time.Hour))
		m.Clear()

		exists, _ := m.Exists(ctx, "clear-test")
		if exists {
			t.Error("expected store to be cleared")
		}
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

		err := m.Set(ctx, "complex", data, time.Now().Add(time.Hour))
		if err != nil {
			t.Fatalf("set failed: %v", err)
		}

		var result nested
		err = m.Get(ctx, "complex", &result)
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
}
