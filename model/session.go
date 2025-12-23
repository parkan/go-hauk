package model

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/parkan/go-hauk/store"
)

const (
	PrefixSession = "session-"
	SessionIDSize = 32
)

type SessionData struct {
	Expire    int64    `json:"expire"`
	Interval  float64  `json:"interval"`
	Targets   []string `json:"targets"`
	Points    [][]any  `json:"points"`
	Encrypted bool     `json:"encrypted"`
	Salt      string   `json:"salt,omitempty"`
}

type Session struct {
	store store.Store
	id    string
	data  SessionData
	maxPts int
}

func NewSession(s store.Store, maxPts int) (*Session, error) {
	id, err := generateSessionID()
	if err != nil {
		return nil, err
	}
	return &Session{
		store:  s,
		id:     id,
		maxPts: maxPts,
		data: SessionData{
			Targets: []string{},
			Points:  [][]any{},
		},
	}, nil
}

func LoadSession(ctx context.Context, s store.Store, id string, maxPts int) (*Session, error) {
	sess := &Session{store: s, id: id, maxPts: maxPts}
	err := s.Get(ctx, PrefixSession+id, &sess.data)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *Session) ID() string               { return s.id }
func (s *Session) Expire() time.Time        { return time.Unix(s.data.Expire, 0) }
func (s *Session) Interval() float64        { return s.data.Interval }
func (s *Session) Targets() []string        { return s.data.Targets }
func (s *Session) Points() [][]any          { return s.data.Points }
func (s *Session) Encrypted() bool          { return s.data.Encrypted }
func (s *Session) Salt() string             { return s.data.Salt }
func (s *Session) HasExpired() bool         { return time.Now().Unix() >= s.data.Expire }

func (s *Session) SetExpire(t time.Time)    { s.data.Expire = t.Unix() }
func (s *Session) SetInterval(i float64)    { s.data.Interval = i }
func (s *Session) SetEncrypted(e bool, salt string) {
	s.data.Encrypted = e
	s.data.Salt = salt
}

func (s *Session) AddTarget(shareID string) {
	s.data.Targets = append(s.data.Targets, shareID)
}

func (s *Session) RemoveTarget(shareID string) {
	for i, t := range s.data.Targets {
		if t == shareID {
			s.data.Targets = append(s.data.Targets[:i], s.data.Targets[i+1:]...)
			return
		}
	}
}

func (s *Session) AddPoint(p []any) {
	s.data.Points = append(s.data.Points, p)
	if len(s.data.Points) > s.maxPts {
		s.data.Points = s.data.Points[len(s.data.Points)-s.maxPts:]
	}
}

func (s *Session) GetPoints(since float64) [][]any {
	if since <= 0 {
		return s.data.Points
	}
	timeIdx := 2
	if s.data.Encrypted {
		timeIdx = 3
	}
	var pts [][]any
	for _, p := range s.data.Points {
		if len(p) > timeIdx {
			if t, ok := p[timeIdx].(float64); ok && t > since {
				pts = append(pts, p)
			}
		}
	}
	return pts
}

func (s *Session) Save(ctx context.Context) error {
	return s.store.Set(ctx, PrefixSession+s.id, s.data, s.Expire())
}

func (s *Session) Delete(ctx context.Context) error {
	return s.store.Delete(ctx, PrefixSession+s.id)
}

func generateSessionID() (string, error) {
	b := make([]byte, SessionIDSize)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
