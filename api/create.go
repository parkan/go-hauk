package api

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/parkan/go-hauk/model"
)

const (
	shareModeCreateAlone = 0
	shareModeCreateGroup = 1
	shareModeJoinGroup   = 2
)

var linkIDRe = regexp.MustCompile(`^[\w-]+$`)

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	dur := r.FormValue("dur")
	interval := r.FormValue("int")
	if dur == "" || interval == "" {
		fmt.Fprintln(w, "Missing data!")
		return
	}

	user := r.FormValue("usr")
	pass := r.FormValue("pwd")
	if err := s.auth.Authenticate(user, pass); err != nil {
		fmt.Fprintln(w, "Incorrect password!")
		return
	}

	d, err := strconv.Atoi(dur)
	if err != nil || d <= 0 {
		fmt.Fprintln(w, "Invalid duration!")
		return
	}
	i, err := strconv.ParseFloat(interval, 64)
	if err != nil || i <= 0 {
		fmt.Fprintln(w, "Invalid interval!")
		return
	}
	mod, _ := strconv.Atoi(r.FormValue("mod"))
	adoptable := r.FormValue("ado") == "1"
	encrypted := r.FormValue("e2e") == "1"
	salt := r.FormValue("salt")
	customLink := r.FormValue("lid")
	nickname := r.FormValue("nic")
	pin, _ := strconv.Atoi(r.FormValue("pin"))

	if d > s.cfg.MaxDuration {
		fmt.Fprintln(w, "Share duration exceeds maximum configured!")
		return
	}
	if i > float64(s.cfg.MaxDuration) {
		fmt.Fprintln(w, "Interval exceeds maximum configured!")
		return
	}
	if i < s.cfg.MinInterval {
		fmt.Fprintln(w, "Interval is too short!")
		return
	}

	if (mod == shareModeCreateGroup || mod == shareModeJoinGroup) && encrypted {
		fmt.Fprintln(w, "End-to-end encryption is not supported for group shares.")
		return
	}

	if (mod == shareModeCreateGroup || mod == shareModeJoinGroup) && nickname == "" {
		fmt.Fprintln(w, "Missing data!")
		return
	}

	if mod == shareModeJoinGroup && pin == 0 {
		fmt.Fprintln(w, "Missing data!")
		return
	}

	if encrypted && salt == "" {
		fmt.Fprintln(w, "Missing data!")
		return
	}

	expire := time.Now().Add(time.Duration(d) * time.Second)

	session, err := model.NewSession(s.store, s.cfg.MaxCachedPts)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session.SetExpire(expire)
	session.SetInterval(i)
	if encrypted {
		session.SetEncrypted(true, salt)
	}

	linkGen := func() (string, error) {
		return s.linkgen.Generate(ctx)
	}

	switch mod {
	case shareModeCreateAlone:
		share, err := model.NewSoloShare(s.store, s.cfg.PublicURL, linkGen)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if customLink != "" && linkIDRe.MatchString(customLink) {
			if err := s.tryCustomLink(ctx, customLink, user); err == nil {
				share.SetID(customLink)
			}
		}

		share.SetAdoptable(adoptable)
		share.SetHost(session.ID())
		share.SetExpire(expire)
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
		fmt.Fprintln(w, session.ID())
		fmt.Fprintln(w, share.ViewLink())
		fmt.Fprintln(w, share.ID())

	case shareModeCreateGroup:
		share, err := model.NewGroupShare(s.store, s.cfg.PublicURL, linkGen)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if customLink != "" && linkIDRe.MatchString(customLink) {
			if err := s.tryCustomLink(ctx, customLink, user); err == nil {
				share.SetID(customLink)
			}
		}

		share.AddHost(nickname, session.ID())
		share.SetExpire(expire)
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
		fmt.Fprintln(w, session.ID())
		fmt.Fprintln(w, share.ViewLink())
		fmt.Fprintln(w, share.Pin())
		fmt.Fprintln(w, share.ID())

	case shareModeJoinGroup:
		share, err := model.LoadGroupShareByPin(ctx, s.store, pin, s.cfg.PublicURL)
		if err != nil {
			fmt.Fprintln(w, "Invalid group PIN!")
			return
		}

		share.AddHost(nickname, session.ID())
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
		fmt.Fprintln(w, session.ID())
		fmt.Fprintln(w, share.ViewLink())
		fmt.Fprintln(w, share.ID())

	default:
		fmt.Fprintln(w, "Unsupported share mode!")
	}
}

func (s *Server) tryCustomLink(ctx context.Context, link, user string) error {
	if !s.cfg.AllowLinkReq {
		return fmt.Errorf("custom links disabled")
	}

	// check reserved links
	if allowedUsers, reserved := s.cfg.ReservedLinks[link]; reserved {
		allowed := false
		for _, u := range allowedUsers {
			if u == user {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("link reserved")
		}
	}

	// whitelist mode: only reserved links allowed
	if s.cfg.ReserveWL {
		if _, reserved := s.cfg.ReservedLinks[link]; !reserved {
			return fmt.Errorf("link not in whitelist")
		}
	}

	exists, err := s.store.Exists(ctx, "locdata-"+link)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("link already exists")
	}
	return nil
}
