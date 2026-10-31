package home

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/raffy-io/taskifyv3/internal/web"
)

func HomeRoutes(sessionManager *scs.SessionManager) http.Handler {
	r := chi.NewRouter()

	renderer := web.NewRenderer(sessionManager)
	h := &homeHandler{
		Renderer: renderer,
	}

	r.Get("/", h.ServeHomepage)

	return r
}
