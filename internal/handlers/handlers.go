package handlers

import (
	"fmt"

	"log/slog"

	"net/http"

	"signflow-2026/internal/config"
)

type Handlers struct {
	Cfg config.Config

	Log *slog.Logger
}

func New(cfg config.Config, log *slog.Logger) *Handlers {

	return &Handlers{Cfg: cfg, Log: log}

}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, `<!doctype html>

<title>SignFlow</title>

<h1>✍️ SignFlow</h1>

<p>The skeleton is up and serving.</p>`)

}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")

	_, _ = w.Write([]byte("ok"))

}
