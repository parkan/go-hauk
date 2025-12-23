package linkgen

import (
	"context"
	"testing"
	"time"

	"github.com/parkan/go-hauk/config"
)

type mockStore struct{}

func (m *mockStore) Get(_ context.Context, _ string, _ any) error { return nil }
func (m *mockStore) Set(_ context.Context, _ string, _ any, _ time.Time) error { return nil }
func (m *mockStore) SetTTL(_ context.Context, _ string, _ any, _ time.Duration) error { return nil }
func (m *mockStore) Delete(_ context.Context, _ string) error { return nil }
func (m *mockStore) Exists(_ context.Context, _ string) (bool, error) { return false, nil }

func TestGenerate4Plus4Upper(t *testing.T) {
	g := New(&mockStore{}, config.Link4Plus4Upper)
	id, err := g.Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 9 {
		t.Errorf("expected 9 chars, got %d: %s", len(id), id)
	}
	if id[4] != '-' {
		t.Errorf("expected dash at position 4: %s", id)
	}
}

func TestGenerate4Plus4Lower(t *testing.T) {
	g := New(&mockStore{}, config.Link4Plus4Lower)
	id, err := g.Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 9 {
		t.Errorf("expected 9 chars, got %d: %s", len(id), id)
	}
}

func TestGenerateUUID(t *testing.T) {
	g := New(&mockStore{}, config.LinkUUIDv4)
	id, err := g.Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 36 {
		t.Errorf("expected 36 chars, got %d: %s", len(id), id)
	}
}

func TestGenerate16Hex(t *testing.T) {
	g := New(&mockStore{}, config.Link16Hex)
	id, err := g.Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 16 {
		t.Errorf("expected 16 chars, got %d: %s", len(id), id)
	}
}

func TestGenerate32Hex(t *testing.T) {
	g := New(&mockStore{}, config.Link32Hex)
	id, err := g.Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 32 {
		t.Errorf("expected 32 chars, got %d: %s", len(id), id)
	}
}
