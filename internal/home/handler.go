package home

import (
	"log/slog"
	"net/http"

	"github.com/raffy-io/taskifyv3/internal/web"
)

type homeHandler struct {
	Renderer *web.Renderer
}

func (h *homeHandler) ServeHomepage(w http.ResponseWriter, r *http.Request) {

	component := homepage()
	if err := h.Renderer.Render(w, r, component); err != nil {
		slog.Error("Failed to render template", "err", err)
		return
	}
}
