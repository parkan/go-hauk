package api

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/parkan/go-hauk/auth"
	"github.com/parkan/go-hauk/config"
	"github.com/parkan/go-hauk/frontend"
	"github.com/parkan/go-hauk/linkgen"
	"github.com/parkan/go-hauk/ratelimit"
	"github.com/parkan/go-hauk/store"
)

const backendVersion = "1.6.2-go"

type Server struct {
	mux         *http.ServeMux
	cfg         *config.Config
	store       store.Store
	auth        auth.Authenticator
	linkgen     *linkgen.Generator
	rlAuth      *ratelimit.Limiter
	rlAdopt     *ratelimit.Limiter
}

func NewServer(cfg *config.Config, s store.Store) *Server {
	srv := &Server{
		mux:     http.NewServeMux(),
		cfg:     cfg,
		store:   s,
		linkgen: linkgen.New(s, cfg.LinkStyle),
		rlAuth:  ratelimit.New(cfg.RateLimitAuth, time.Minute, cfg.TrustProxy),
		rlAdopt: ratelimit.New(cfg.RateLimitAdopt, time.Minute, cfg.TrustProxy),
	}

	switch cfg.AuthMethod {
	case config.AuthHtpasswd:
		srv.auth = auth.NewHtpasswdAuth(cfg.HtpasswdPath)
	case config.AuthLDAP:
		srv.auth = auth.NewLDAPAuth(
			cfg.LDAPUri, cfg.LDAPBaseDN, cfg.LDAPBindDN,
			cfg.LDAPBindPass, cfg.LDAPUserFilter, cfg.LDAPStartTLS,
		)
	default:
		srv.auth = auth.NewPasswordAuth(cfg.PasswordHash)
	}

	srv.mux.HandleFunc("POST /api/create.php", srv.rlAuth.WrapFunc(srv.handleCreate))
	srv.mux.HandleFunc("POST /api/post.php", srv.handlePost)
	srv.mux.HandleFunc("GET /api/fetch.php", srv.handleFetch)
	srv.mux.HandleFunc("POST /api/stop.php", srv.handleStop)
	srv.mux.HandleFunc("POST /api/adopt.php", srv.rlAdopt.WrapFunc(srv.handleAdopt))
	srv.mux.HandleFunc("POST /api/new-link.php", srv.handleNewLink)
	srv.mux.HandleFunc("GET /dynamic.js.php", srv.handleDynamic)

	staticFS, _ := fs.Sub(frontend.Files, ".")
	srv.mux.Handle("/", http.FileServer(http.FS(staticFS)))

	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Hauk-Version", backendVersion)
	s.mux.ServeHTTP(w, r)
}
