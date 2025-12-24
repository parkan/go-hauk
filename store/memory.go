package store

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("key not found")

type entry struct {
	data     []byte
	expireAt time.Time
}

// Memory is an in-memory store for testing
type Memory struct {
	mu   sync.RWMutex
	data map[string]entry
}

func NewMemory() *Memory {
	return &Memory{
		data: make(map[string]entry),
	}
}

func (m *Memory) Get(ctx context.Context, key string, v any) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	e, ok := m.data[key]
	if !ok {
		return ErrNotFound
	}
	if !e.expireAt.IsZero() && time.Now().After(e.expireAt) {
		return ErrNotFound
	}
	return json.Unmarshal(e.data, v)
}

func (m *Memory) Set(ctx context.Context, key string, v any, expireAt time.Time) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = entry{data: data, expireAt: expireAt}
	return nil
}

func (m *Memory) SetTTL(ctx context.Context, key string, v any, ttl time.Duration) error {
	return m.Set(ctx, key, v, time.Now().Add(ttl))
}

func (m *Memory) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

func (m *Memory) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	e, ok := m.data[key]
	if !ok {
		return false, nil
	}
	if !e.expireAt.IsZero() && time.Now().After(e.expireAt) {
		return false, nil
	}
	return true, nil
}

func (m *Memory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]entry)
}
