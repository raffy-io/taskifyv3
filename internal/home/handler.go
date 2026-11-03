package home

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/raffy-io/taskifyv3/internal/db"
	"github.com/raffy-io/taskifyv3/internal/imageutil"
	"github.com/raffy-io/taskifyv3/internal/web"
)

type homeHandler struct {
	Renderer *web.Renderer
	Queries  *db.Queries
	Session  *scs.SessionManager
	Cld      *cloudinary.Cloudinary
}

func newHomeHandler(renderer *web.Renderer, queries *db.Queries, session *scs.SessionManager, cld *cloudinary.Cloudinary) *homeHandler {
	return &homeHandler{
		Renderer: renderer,
		Queries:  queries,
		Session:  session,
		Cld:      cld,
	}
}

func (h *homeHandler) ServeHomepage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	homedata, err := h.Queries.GetHomepage(ctx, 1)
	if err != nil {
		http.Error(w, "No data for homepage yet", http.StatusInternalServerError)
		return
	}
	component := homepage(homedata)
	if err := h.Renderer.Render(w, r, component); err != nil {
		slog.Error("Failed to render template", "err", err)
		return
	}
}

func (h *homeHandler) ServeEditHomepage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	existing, err := h.Queries.GetHomepage(ctx, 1)
	if err != nil {
		// If the table is completely empty, initialize with empty/default values
		// instead of throwing a 500 error.
		form := newEditHomepageForm("", "", "", "")
		component := homepageEditor(form)
		if renderErr := h.Renderer.Render(w, r, component); renderErr != nil {
			slog.Error("Failed to render template", "err", renderErr)
		}
		return
	}

	form := newEditHomepageForm(existing.Title, existing.Subtitle, existing.HomeImage, existing.HomeImageAlt)
	component := homepageEditor(form)
	if err := h.Renderer.Render(w, r, component); err != nil {
		slog.Error("Failed to render template", "err", err)
		return
	}
}

const maxUploadSize = 10 << 20 // 10 MB
func (h *homeHandler) ProcessUpdateHomepage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(15 << 20); err != nil {
		http.Error(w, "The uploaded file is too big", http.StatusRequestEntityTooLarge)
		return
	}
	defer r.MultipartForm.RemoveAll()

	existing, err := h.Queries.GetHomepage(ctx, 1)
	var existingImage string
	if err == nil {
		existingImage = existing.HomeImage
	}
	// Now use existingImage instead of existing.HomeImage when no file is uploaded

	title := r.PostFormValue("title")
	subtitle := r.PostFormValue("subtitle")
	imageAlt := r.PostFormValue("home_image_alt")

	// Default to the existing image path if no new file is uploaded
	imagePath := existingImage

	file, header, err := r.FormFile("home_image")
	if err == nil {
		defer file.Close()

		if valErr := imageutil.ValidateImage(file, header.Filename, header.Size, maxUploadSize); valErr != nil {
			form := newEditHomepageForm(title, subtitle, existing.HomeImage, imageAlt)
			form.addError("homeImage", valErr.Error())

			w.WriteHeader(http.StatusUnprocessableEntity)
			if err := homepageEditor(form).Render(ctx, w); err != nil {
				slog.Error("Failed to render template", "err", err)
			}
			return
		}

		webpImage, err := imageutil.ProcessImage(file, header.Filename, 80)
		if err != nil {
			http.Error(w, "Failed to process image", http.StatusInternalServerError)
			return
		}

		// --- UPLOAD TO CLOUDINARY INSTEAD OF LOCAL DISK ---
		// We pass the raw webpImage bytes using bytes.NewReader
		uploadResult, err := h.Cld.Upload.Upload(
			ctx,
			bytes.NewReader(webpImage),
			uploader.UploadParams{
				Folder:   "taskify/hero", // Optional: organizes files in a folder on your Cloudinary dashboard
				PublicID: "hero_image",   // Optional: fixed name so it overwrites/updates cleanly, or omit for random hash
			},
		)
		if err != nil {
			slog.Error("Failed to upload image to Cloudinary", "err", err)
			http.Error(w, "Failed to upload image to cloud", http.StatusInternalServerError)
			return
		}

		// Assign the secure HTTPS cloud URL returned by Cloudinary
		imagePath = uploadResult.SecureURL

	} else if !errors.Is(err, http.ErrMissingFile) {
		form := newEditHomepageForm(title, subtitle, existing.HomeImage, imageAlt)
		form.addError("homeImage", "Failed to process uploaded image.")
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := homepageEditor(form).Render(ctx, w); err != nil {
			slog.Error("Failed to render template", "err", err)
		}
		return
	}

	// Now validate the full form including the final imagePath (if needed)
	form := newEditHomepageForm(title, subtitle, imagePath, imageAlt)
	if !form.validate() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := homepageEditor(form).Render(ctx, w); err != nil {
			slog.Error("Failed to render template", "err", err)
		}
		return
	}

	// TODO: Execute your SQLC update query here using title, subtitle, imagePath, imageAlt...
	err = h.Queries.UpdateHomepage(ctx, db.UpdateHomepageParams{
		ID:           1,
		Title:        form.Title,
		Subtitle:     form.Subtitle,
		HomeImage:    imagePath,
		HomeImageAlt: imageAlt,
	})
	if err != nil {
		slog.Error("failed to update homepage", "err", err.Error())
		http.Error(w, "failed to update homepage", http.StatusInternalServerError)
	}

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}
