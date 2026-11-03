package home

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/raffy-io/taskifyv3/internal/db"
	"github.com/raffy-io/taskifyv3/internal/web"
)

func HomeRoutes(queries *db.Queries, sessionManager *scs.SessionManager, cld *cloudinary.Cloudinary) http.Handler {
	r := chi.NewRouter()

	renderer := web.NewRenderer(sessionManager)
	h := newHomeHandler(renderer, queries, sessionManager, cld)

	r.Get("/", h.ServeHomepage)

	return r
}

func AdminRoutes(queries *db.Queries, sessionManager *scs.SessionManager, cld *cloudinary.Cloudinary) http.Handler {
	r := chi.NewRouter()

	renderer := web.NewRenderer(sessionManager)
	h := newHomeHandler(renderer, queries, sessionManager, cld)

	r.Get("/home/edit", h.ServeEditHomepage)
	r.Patch("/home/edit", h.ProcessUpdateHomepage)

	return r
}
