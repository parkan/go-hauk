package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/parkan/go-hauk/model"
)

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	sid := r.FormValue("sid")
	lat := r.FormValue("lat")
	lon := r.FormValue("lon")
	ts := r.FormValue("time")

	if sid == "" || lat == "" || lon == "" || ts == "" {
		fmt.Fprintln(w, "Missing data!")
		return
	}

	session, err := model.LoadSession(ctx, s.store, sid, s.cfg.MaxCachedPts)
	if err != nil {
		fmt.Fprintln(w, "Session expired!")
		return
	}

	var point []any

	if !session.Encrypted() {
		latF, err1 := strconv.ParseFloat(lat, 64)
		lonF, err2 := strconv.ParseFloat(lon, 64)
		timeF, err3 := strconv.ParseFloat(ts, 64)

		if err1 != nil || err2 != nil || err3 != nil ||
			latF < -90 || latF > 90 || lonF < -180 || lonF > 180 {
			fmt.Fprintln(w, "Invalid location!")
			return
		}

		var speed, acc *float64
		if spd := r.FormValue("spd"); spd != "" {
			v, _ := strconv.ParseFloat(spd, 64)
			speed = &v
		}
		if a := r.FormValue("acc"); a != "" {
			v, _ := strconv.ParseFloat(a, 64)
			acc = &v
		}

		prv := 0
		if r.FormValue("prv") == "1" {
			prv = 1
		}

		point = []any{latF, lonF, timeF, prv, acc, speed}
	} else {
		iv := r.FormValue("iv")
		if iv == "" {
			fmt.Fprintln(w, "Missing data!")
			return
		}

		var speed, acc, prv any
		if spd := r.FormValue("spd"); spd != "" {
			speed = spd
		}
		if a := r.FormValue("acc"); a != "" {
			acc = a
		}
		if p := r.FormValue("prv"); p != "" {
			prv = p
		}

		point = []any{iv, lat, lon, ts, prv, acc, speed}
	}

	session.AddPoint(point)
	if err := session.Save(ctx); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if session.HasExpired() {
		fmt.Fprintln(w, "Session expired!")
		return
	}

	fmt.Fprintln(w, "OK")
	fmt.Fprintf(w, "%s?%%s\n", s.cfg.PublicURL)
	fmt.Fprintln(w, strings.Join(session.Targets(), ","))
}
