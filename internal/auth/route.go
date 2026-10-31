package auth

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/raffy-io/taskifyv3/internal/db"
)

func AuthRoutes(queries *db.Queries, sessionManager *scs.SessionManager) http.Handler {
	r := chi.NewRouter()
	
	h := &authHandler{
		Queries: queries,
		Session: sessionManager,
	}

	r.Get("/login",h.ServeLoginPage)
	r.Get("/register",h.ServeRegisterPage)
	r.Post("/register", h.processRegister)
	r.Post("/login", h.processLogin)
	r.Delete("/logout", h.processLogout)

	return r
}
