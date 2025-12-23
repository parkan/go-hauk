package model

import (
	"testing"
	"time"
)

func TestSessionAddPoint(t *testing.T) {
	sess := &Session{
		maxPts: 3,
		data: SessionData{
			Points: [][]any{},
		},
	}

	for i := 0; i < 5; i++ {
		sess.AddPoint([]any{float64(i), float64(i), float64(i), 0, nil, nil})
	}

	if len(sess.data.Points) != 3 {
		t.Errorf("expected 3 points, got %d", len(sess.data.Points))
	}

	if sess.data.Points[0][0].(float64) != 2 {
		t.Errorf("expected oldest point to be 2, got %v", sess.data.Points[0][0])
	}
}

func TestSessionGetPoints(t *testing.T) {
	sess := &Session{
		maxPts: 10,
		data: SessionData{
			Points: [][]any{
				{1.0, 1.0, 100.0, 0, nil, nil},
				{2.0, 2.0, 200.0, 0, nil, nil},
				{3.0, 3.0, 300.0, 0, nil, nil},
			},
		},
	}

	pts := sess.GetPoints(150)
	if len(pts) != 2 {
		t.Errorf("expected 2 points after since=150, got %d", len(pts))
	}
}

func TestSessionExpired(t *testing.T) {
	sess := &Session{
		data: SessionData{
			Expire: time.Now().Add(-time.Hour).Unix(),
		},
	}

	if !sess.HasExpired() {
		t.Error("expected session to be expired")
	}

	sess.data.Expire = time.Now().Add(time.Hour).Unix()
	if sess.HasExpired() {
		t.Error("expected session to not be expired")
	}
}
