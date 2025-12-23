package api

import (
	"fmt"
	"net/http"

	"github.com/parkan/go-hauk/model"
)

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	sid := r.FormValue("sid")
	if sid == "" {
		fmt.Fprintln(w, "OK")
		return
	}

	lid := r.FormValue("lid")

	session, err := model.LoadSession(ctx, s.store, sid, s.cfg.MaxCachedPts)
	if err != nil {
		fmt.Fprintln(w, "OK")
		return
	}

	if lid != "" {
		found := false
		for _, t := range session.Targets() {
			if t == lid {
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintln(w, "OK")
			return
		}

		shareType, err := model.LoadShareType(ctx, s.store, lid)
		if err == nil {
			switch shareType {
			case model.ShareTypeAlone:
				share, err := model.LoadSoloShare(ctx, s.store, lid, s.cfg.PublicURL)
				if err == nil {
					share.Delete(ctx)
				}
			case model.ShareTypeGroup:
				share, err := model.LoadGroupShare(ctx, s.store, lid, s.cfg.PublicURL)
				if err == nil {
					share.RemoveHost(sid)
					if len(share.Hosts()) == 0 {
						share.Delete(ctx)
					} else {
						share.Save(ctx)
					}
				}
			}
		}

		session.RemoveTarget(lid)
		session.Save(ctx)
	} else {
		for _, t := range session.Targets() {
			shareType, err := model.LoadShareType(ctx, s.store, t)
			if err != nil {
				continue
			}
			switch shareType {
			case model.ShareTypeAlone:
				share, err := model.LoadSoloShare(ctx, s.store, t, s.cfg.PublicURL)
				if err == nil {
					share.Delete(ctx)
				}
			case model.ShareTypeGroup:
				share, err := model.LoadGroupShare(ctx, s.store, t, s.cfg.PublicURL)
				if err == nil {
					share.RemoveHost(sid)
					if len(share.Hosts()) == 0 {
						share.Delete(ctx)
					} else {
						share.Save(ctx)
					}
				}
			}
		}
		session.Delete(ctx)
	}

	fmt.Fprintln(w, "OK")
}
