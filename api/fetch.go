package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/parkan/go-hauk/model"
)

type soloResponse struct {
	Type       int       `json:"type"`
	Expire     int64     `json:"expire"`
	ServerTime float64   `json:"serverTime"`
	Interval   float64   `json:"interval"`
	Points     [][]any   `json:"points"`
	Encrypted  bool      `json:"encrypted"`
	Salt       string    `json:"salt"`
}

type groupResponse struct {
	Type       int                 `json:"type"`
	Expire     int64               `json:"expire"`
	ServerTime float64             `json:"serverTime"`
	Interval   float64             `json:"interval"`
	Points     map[string][][]any  `json:"points"`
}

func (s *Server) handleFetch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.URL.Query().Get("id")
	if id == "" {
		fmt.Fprintln(w, "Invalid session!")
		return
	}

	sinceStr := r.URL.Query().Get("since")
	since := float64(0)
	if sinceStr != "" {
		since, _ = strconv.ParseFloat(sinceStr, 64)
	}

	shareType, err := model.LoadShareType(ctx, s.store, id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Invalid session!")
		return
	}

	w.Header().Set("Content-Type", "text/json")
	now := float64(time.Now().UnixNano()) / 1e9

	switch shareType {
	case model.ShareTypeAlone:
		share, err := model.LoadSoloShare(ctx, s.store, id, s.cfg.PublicURL)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Invalid session!")
			return
		}

		session, err := model.LoadSession(ctx, s.store, share.Host(), s.cfg.MaxCachedPts)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Invalid session!")
			return
		}

		resp := soloResponse{
			Type:       share.Type(),
			Expire:     share.Expire().Unix(),
			ServerTime: now,
			Interval:   session.Interval(),
			Points:     session.GetPoints(since),
			Encrypted:  session.Encrypted(),
			Salt:       session.Salt(),
		}
		if resp.Points == nil {
			resp.Points = [][]any{}
		}
		json.NewEncoder(w).Encode(resp)

	case model.ShareTypeGroup:
		share, err := model.LoadGroupShare(ctx, s.store, id, s.cfg.PublicURL)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Invalid session!")
			return
		}

		points, _ := share.GetAllPoints(ctx, since, s.cfg.MaxCachedPts)
		if points == nil {
			points = make(map[string][][]any)
		}

		resp := groupResponse{
			Type:       share.Type(),
			Expire:     share.Expire().Unix(),
			ServerTime: now,
			Interval:   share.GetAutoInterval(ctx, s.cfg.MaxCachedPts),
			Points:     points,
		}
		json.NewEncoder(w).Encode(resp)
	}
}
