package handlers

import (
	"log/slog"
	"net/http"
	"signflow-2026/internal/config"
	"signflow-2026/internal/db"
	"signflow-2026/internal/email"
	"signflow-2026/internal/session"
	"signflow-2026/internal/web"
)

type Handlers struct {
	Cfg      config.Config
	Queries  *db.Queries
	Sessions *session.Manager
	Mailer   email.Sender
	Log      *slog.Logger
}

func New(cfg config.Config, q *db.Queries, sessions *session.Manager, mailer email.Sender, log *slog.Logger) *Handlers {
	return &Handlers{Cfg: cfg, Queries: q, Sessions: sessions, Mailer: mailer, Log: log}
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	if _, ok := UserFrom(r.Context()); ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	count, err := h.Queries.CountUsers(r.Context())
	if err != nil {
		h.serverError(w, r, "count users", err)
		return
	}
	render(w, r, http.StatusOK, web.Home(h.nav(r), count))
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	if _, err := h.Queries.CountUsers(r.Context()); err != nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("ok"))
}

func (h *Handlers) serverError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	h.Log.Error(msg, "err", err, "path", r.URL.Path)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
