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

	// verify caller owns the share being adopted
	// after this check, session IS the host session
	if sid != share.Host() {
		fmt.Fprintln(w, "Not authorized!")
		return
	}

	if session.Encrypted() {
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

	session.AddTarget(target.ID())
	if err := session.Save(ctx); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "OK")
}
