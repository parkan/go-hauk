package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/parkan/go-hauk/model"
)

func (s *Server) handleAdopt(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	sid := r.FormValue("sid")
	nickname := r.FormValue("nic")
	aid := r.FormValue("aid")
	pinStr := r.FormValue("pin")

	if sid == "" || nickname == "" || aid == "" || pinStr == "" {
		fmt.Fprintln(w, "Missing data!")
		return
	}

	session, err := model.LoadSession(ctx, s.store, sid, s.cfg.MaxCachedPts)
	if err != nil {
		fmt.Fprintln(w, "Session expired!")
		return
	}

	shareType, err := model.LoadShareType(ctx, s.store, aid)
	if err != nil {
		fmt.Fprintln(w, "Share not found!")
		return
	}

	if shareType != model.ShareTypeAlone {
		fmt.Fprintln(w, "Group shares cannot be adopted!")
		return
	}

	share, err := model.LoadSoloShare(ctx, s.store, aid, s.cfg.PublicURL)
	if err != nil {
		fmt.Fprintln(w, "Share not found!")
		return
	}

	if !share.Adoptable() {
		fmt.Fprintln(w, "Share adoption not allowed!")
		return
	}

	hostSession, err := model.LoadSession(ctx, s.store, share.Host(), s.cfg.MaxCachedPts)
	if err != nil {
		fmt.Fprintln(w, "Session expired!")
		return
	}

	if hostSession.Encrypted() {
		fmt.Fprintln(w, "End-to-end encrypted shares cannot be adopted!")
		return
	}

	pin, _ := strconv.Atoi(pinStr)
	target, err := model.LoadGroupShareByPin(ctx, s.store, pin, s.cfg.PublicURL)
	if err != nil {
		fmt.Fprintln(w, "Session expired!")
		return
	}

	target.AddHost(nickname, share.Host())
	if err := target.Save(ctx); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	hostSession.AddTarget(target.ID())
	if err := hostSession.Save(ctx); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = session

	fmt.Fprintln(w, "OK")
}
