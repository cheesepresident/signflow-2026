package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"signflow-2026/internal/config"
	"signflow-2026/internal/handlers"
	"signflow-2026/static"
)

func main() {

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := run(log); err != nil {

		log.Error("fatal", "err", err)

		os.Exit(1)

	}

}

func run(log *slog.Logger) error {

	cfg, err := config.Load()

	if err != nil {

		return err

	}

	log.Info("starting signflow", "env", cfg.Env, "port", cfg.Port, "base_url", cfg.BaseURL)

	h := handlers.New(cfg, log)

	staticFS, err := fs.Sub(static.FS, "assets")
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           h.Router(staticFS),
		ReadHeaderTimeout: 10 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	go func() {

		log.Info("listening", "addr", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {

			log.Error("server error", "err", err)

			stop()

		}

	}()

	<-shutdownCtx.Done()

	log.Info("shutting down")

	timeout, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	return srv.Shutdown(timeout)

}
