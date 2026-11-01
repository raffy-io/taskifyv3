package auth

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/raffy-io/taskifyv3/internal/db"
	"github.com/raffy-io/taskifyv3/internal/web"
)

func AuthRoutes(queries *db.Queries, sessionManager *scs.SessionManager) http.Handler {
	r := chi.NewRouter()

	renderer := web.NewRenderer(sessionManager)
	h := newAuthHandler(queries, sessionManager, renderer)

	r.Get("/login", h.ServeLoginPage)
	r.Get("/register", h.ServeRegisterPage)
	r.Post("/register", h.processRegister)
	r.Post("/login", h.processLogin)
	r.Delete("/logout", h.processLogout)

	return r
}
