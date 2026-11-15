package handlers

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

func (h *Handlers) Router(staticFS fs.FS, csrfKey []byte) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)    // request logging
	r.Use(middleware.Recoverer) // turn panics into 500s instead of crashing

	if !h.Cfg.IsProd() {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				next.ServeHTTP(w, csrf.PlaintextHTTPRequest(req))
			})
		})
	}

	csrfMW := csrf.Protect(
		csrfKey,
		csrf.Secure(h.Cfg.IsProd()),
		csrf.Path("/"),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.FieldName("gorilla.csrf.Token"),
	)
	r.Use(csrfMW)

	r.Use(h.LoadUser)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	r.Get("/", h.Home)
	r.Get("/healthz", h.Health)

	r.Get("/register", h.RegisterForm)
	r.Post("/register", h.Register)
	r.Get("/login", h.LoginForm)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Get("/forgot", h.ForgotForm)
	r.Post("/forgot", h.Forgot)
	r.Get("/reset/{token}", h.ResetForm)
	r.Post("/reset/{token}", h.Reset)

	r.Group(func(pr chi.Router) {
		pr.Use(h.RequireAuth)
		pr.Get("/dashboard", h.Dashboard)
	})

	return r
}
