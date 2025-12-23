package api

import (
	"fmt"
	"net/http"

	"github.com/parkan/go-hauk/model"
)

func (s *Server) handleNewLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	sid := r.FormValue("sid")
	adoptable := r.FormValue("ado") == "1"

	if sid == "" {
		fmt.Fprintln(w, "Missing data!")
		return
	}

	session, err := model.LoadSession(ctx, s.store, sid, s.cfg.MaxCachedPts)
	if err != nil {
		fmt.Fprintln(w, "Session expired!")
		return
	}

	linkGen := func() (string, error) {
		return s.linkgen.Generate(ctx)
	}

	share, err := model.NewSoloShare(s.store, s.cfg.PublicURL, linkGen)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	share.SetAdoptable(adoptable)
	share.SetHost(session.ID())
	share.SetExpire(session.Expire())

	if err := share.Save(ctx); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	session.AddTarget(share.ID())
	if err := session.Save(ctx); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "OK")
	fmt.Fprintln(w, share.ViewLink())
	fmt.Fprintln(w, share.ID())
}
