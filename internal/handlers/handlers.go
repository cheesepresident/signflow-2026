package handlers

import (
	"log/slog"

	"net/http"

	"signflow-2026/internal/config"
	"signflow-2026/internal/web"
)

type Handlers struct {
	Cfg config.Config

	Log *slog.Logger
}

func New(cfg config.Config, log *slog.Logger) *Handlers {

	return &Handlers{Cfg: cfg, Log: log}

}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	render(w, r, http.StatusOK, web.Home())
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")

	_, _ = w.Write([]byte("ok"))

}
