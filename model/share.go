package model

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"strconv"
	"time"

	"github.com/parkan/go-hauk/store"
)

const (
	PrefixLocdata = "locdata-"
	PrefixGroupID = "groupid-"

	ShareTypeAlone = 0
	ShareTypeGroup = 1

	GroupPinMin = 100000
	GroupPinMax = 999999
)

type SoloShareData struct {
	Type      int    `json:"type"`
	Expire    int64  `json:"expire"`
	Host      string `json:"host"`
	Adoptable bool   `json:"adoptable"`
}

type GroupShareData struct {
	Type     int               `json:"type"`
	Expire   int64             `json:"expire"`
	Hosts    map[string]string `json:"hosts"`
	GroupPin int               `json:"groupPin"`
}

type SoloShare struct {
	store     store.Store
	id        string
	data      SoloShareData
	publicURL string
}

type GroupShare struct {
	store     store.Store
	id        string
	data      GroupShareData
	publicURL string
}

func NewSoloShare(s store.Store, publicURL string, linkGen func() (string, error)) (*SoloShare, error) {
	id, err := linkGen()
	if err != nil {
		return nil, err
	}
	return &SoloShare{
		store:     s,
		id:        id,
		publicURL: publicURL,
		data: SoloShareData{
			Type: ShareTypeAlone,
		},
	}, nil
}

func LoadSoloShare(ctx context.Context, s store.Store, id, publicURL string) (*SoloShare, error) {
	share := &SoloShare{store: s, id: id, publicURL: publicURL}
	err := s.Get(ctx, PrefixLocdata+id, &share.data)
	if err != nil {
		return nil, err
	}
	return share, nil
}

func (s *SoloShare) ID() string             { return s.id }
func (s *SoloShare) Type() int              { return s.data.Type }
func (s *SoloShare) Expire() time.Time      { return time.Unix(s.data.Expire, 0) }
func (s *SoloShare) Host() string           { return s.data.Host }
func (s *SoloShare) Adoptable() bool        { return s.data.Adoptable }
func (s *SoloShare) ViewLink() string       { return s.publicURL + "?" + s.id }

func (s *SoloShare) SetExpire(t time.Time)  { s.data.Expire = t.Unix() }
func (s *SoloShare) SetHost(sid string)     { s.data.Host = sid }
func (s *SoloShare) SetAdoptable(a bool)    { s.data.Adoptable = a }
func (s *SoloShare) SetID(id string)        { s.id = id }

func (s *SoloShare) Save(ctx context.Context) error {
	return s.store.Set(ctx, PrefixLocdata+s.id, s.data, s.Expire())
}

func (s *SoloShare) Delete(ctx context.Context) error {
	return s.store.Delete(ctx, PrefixLocdata+s.id)
}

func NewGroupShare(s store.Store, publicURL string, linkGen func() (string, error)) (*GroupShare, error) {
	id, err := linkGen()
	if err != nil {
		return nil, err
	}
	pin := GroupPinMin + cryptoRandInt(GroupPinMax-GroupPinMin+1)
	return &GroupShare{
		store:     s,
		id:        id,
		publicURL: publicURL,
		data: GroupShareData{
			Type:     ShareTypeGroup,
			Hosts:    make(map[string]string),
			GroupPin: pin,
		},
	}, nil
}

func LoadGroupShare(ctx context.Context, s store.Store, id, publicURL string) (*GroupShare, error) {
	share := &GroupShare{store: s, id: id, publicURL: publicURL}
	err := s.Get(ctx, PrefixLocdata+id, &share.data)
	if err != nil {
		return nil, err
	}
	return share, nil
}

func LoadGroupShareByPin(ctx context.Context, s store.Store, pin int, publicURL string) (*GroupShare, error) {
	var shareID string
	err := s.Get(ctx, PrefixGroupID+strconv.Itoa(pin), &shareID)
	if err != nil {
		return nil, err
	}
	return LoadGroupShare(ctx, s, shareID, publicURL)
}

func (g *GroupShare) ID() string              { return g.id }
func (g *GroupShare) Type() int               { return g.data.Type }
func (g *GroupShare) Expire() time.Time       { return time.Unix(g.data.Expire, 0) }
func (g *GroupShare) Hosts() map[string]string { return g.data.Hosts }
func (g *GroupShare) Pin() int                { return g.data.GroupPin }
func (g *GroupShare) ViewLink() string        { return g.publicURL + "?" + g.id }

func (g *GroupShare) SetExpire(t time.Time)   { g.data.Expire = t.Unix() }
func (g *GroupShare) SetID(id string)         { g.id = id }

func (g *GroupShare) AddHost(nick, sessionID string) {
	g.data.Hosts[nick] = sessionID
}

func (g *GroupShare) RemoveHost(sessionID string) {
	for nick, sid := range g.data.Hosts {
		if sid == sessionID {
			delete(g.data.Hosts, nick)
			return
		}
	}
}

func (g *GroupShare) Save(ctx context.Context) error {
	if err := g.store.Set(ctx, PrefixLocdata+g.id, g.data, g.Expire()); err != nil {
		return err
	}
	return g.store.Set(ctx, PrefixGroupID+strconv.Itoa(g.data.GroupPin), g.id, g.Expire())
}

func (g *GroupShare) Delete(ctx context.Context) error {
	g.store.Delete(ctx, PrefixGroupID+strconv.Itoa(g.data.GroupPin))
	return g.store.Delete(ctx, PrefixLocdata+g.id)
}

func (g *GroupShare) GetAllPoints(ctx context.Context, since float64, maxPts int) (map[string][][]any, error) {
	points := make(map[string][][]any)
	for nick, sid := range g.data.Hosts {
		sess, err := LoadSession(ctx, g.store, sid, maxPts)
		if err != nil {
			continue
		}
		points[nick] = sess.GetPoints(since)
	}
	return points, nil
}

func (g *GroupShare) GetAutoInterval(ctx context.Context, maxPts int) float64 {
	min := float64(0)
	for _, sid := range g.data.Hosts {
		sess, err := LoadSession(ctx, g.store, sid, maxPts)
		if err != nil {
			continue
		}
		if min == 0 || sess.Interval() < min {
			min = sess.Interval()
		}
	}
	return min
}

type ShareType struct {
	Type int `json:"type"`
}

func LoadShareType(ctx context.Context, s store.Store, id string) (int, error) {
	var st ShareType
	err := s.Get(ctx, PrefixLocdata+id, &st)
	if err != nil {
		return -1, err
	}
	return st.Type, nil
}

func cryptoRandInt(max int) int {
	var b [8]byte
	rand.Read(b[:])
	return int(binary.LittleEndian.Uint64(b[:]) % uint64(max))
}
