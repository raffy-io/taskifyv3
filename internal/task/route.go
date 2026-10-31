package task

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/raffy-io/taskifyv3/internal/db"
)

func TaskRoutes(queries *db.Queries,sessionManager *scs.SessionManager) http.Handler{
	r := chi.NewRouter()

	h := &taskHandler{
		Queries: queries,
		Session: sessionManager,
	}

	r.Post("/",h.createTask)
	r.Patch("/{id}/toggle",h.toggleTask)
	r.Delete("/{id}",h.deleteTask)

	return r

}