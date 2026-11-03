package auth

import (
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/raffy-io/taskifyv3/internal/db"
	"github.com/raffy-io/taskifyv3/internal/utils"
	"github.com/raffy-io/taskifyv3/internal/web"
)

type authHandler struct {
	Queries  *db.Queries
	Session  *scs.SessionManager
	Renderer *web.Renderer
}

// Initialize the auth handler struct
func newAuthHandler(queries *db.Queries, session *scs.SessionManager, renderer *web.Renderer) *authHandler {
	return &authHandler{
		Queries:  queries,
		Session:  session,
		Renderer: renderer,
	}
}

// authHandler Methods

func (h *authHandler) ServeRegisterPage(w http.ResponseWriter, r *http.Request) {
	form := newRegisterForm("", "")
	component := registerFormPage(form)
	if err := h.Renderer.Render(w, r, component); err != nil {
		slog.Error("Failed to render template", "err", err)
		return
	}

}
func (h *authHandler) ServeLoginPage(w http.ResponseWriter, r *http.Request) {
	form := newLoginForm("")
	component := loginFormPage(form)
	if err := h.Renderer.Render(w, r, component); err != nil {
		slog.Error("Failed to render template", "err", err)
		return
	}
}

func (h *authHandler) processRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// 3. Build the clean form struct
	form := newRegisterForm(r.PostFormValue("username"), r.PostFormValue("email"))

	// separate validation for password to avoid leaving info in frontend
	password := r.PostFormValue("password") // Never lowercase passwords!
	if !form.validate(password) {
		// 1. Return 422 UnprocessableEntity for validation failures
		w.WriteHeader(http.StatusUnprocessableEntity)
		// 2. Render the Templ component with the form errors attached
		if err := registerFormPage(form).Render(ctx, w); err != nil {
			slog.Error("Failed to render template", "err", err)
		}
		return
	}
	// 2. Syntax passed! Now check database uniqueness using SQLC
	usernameExists, err := h.Queries.CheckUsernameExists(ctx, form.Username)
	if err != nil {
		slog.Error("Failed to check user uniqueness", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if usernameExists {
		form.addError("username", "This username is already taken.")

		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := registerFormPage(form).Render(ctx, w); err != nil {
			slog.Error("Failed to render template", "err", err)
		}
		return
	}

	// 2. Syntax passed! Now check database uniqueness using SQLC
	emailExists, err := h.Queries.CheckEmailExists(ctx, form.Email)
	if err != nil {
		slog.Error("Failed to check email uniqueness", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if emailExists {
		form.addError("email", "This email address is already registered.")

		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := registerFormPage(form).Render(ctx, w); err != nil {
			slog.Error("Failed to render template", "err", err)
		}
		return
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		http.Error(w, "Invalid Password", http.StatusInternalServerError)
		return
	}

	user, err := h.Queries.CreateUser(ctx, db.CreateUserParams{
		Username:     form.Username,
		Email:        form.Email,
		PasswordHash: hashedPassword,
	})

	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	slog.Info("Succesfully registered", "user", user.Username)
	// 3. Renew session token (prevents session fixation)
	if err := h.Session.RenewToken(ctx); err != nil {
		slog.Error("Failed to renew session token", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 4. Set session user_id
	h.Session.Put(ctx, "user_id", user.ID)
	h.Session.Put(ctx, "role", user.Role)

	// 5. Redirect straight to app dashboard
	w.Header().Set("HX-Location", "/task")
	w.WriteHeader(http.StatusOK)
}
func (h *authHandler) processLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Initialize the form with submitted values and an empty Errors map
	form := newLoginForm(r.PostFormValue("email"))
	password := r.PostFormValue("password")
	if !form.validate(password) {
		// 1. Return 422 UnprocessableEntity for validation failures
		w.WriteHeader(http.StatusUnprocessableEntity)
		// 2. Render the Templ component with the form errors attached
		if err := loginFormPage(form).Render(ctx, w); err != nil {
			slog.Error("Failed to render template", "err", err)
		}
		// 3. Stop execution so it doesn't proceed to user creation!
		return
	}

	// 1. Check if email exists
	user, err := h.Queries.GetUserByEmail(ctx, form.Email)
	if err != nil {
		form.addError("email", "email not found")
		// 1. Return 422 UnprocessableEntity for validation failures
		w.WriteHeader(http.StatusUnprocessableEntity)
		if renderErr := loginFormPage(form).Render(ctx, w); renderErr != nil {
			slog.Error("Failed to render login template", "err", renderErr)
		}
		return
	}

	// 2. Check password
	if !utils.CheckPassword(password, user.PasswordHash) {
		form.Errors["password"] = "wrong password"
		// Render template with 200 OK so HTMX swaps the form back in
		w.WriteHeader(http.StatusUnprocessableEntity)
		if renderErr := loginFormPage(form).Render(ctx, w); renderErr != nil {
			slog.Error("Failed to render login template", "err", renderErr)
		}
		return
	}

	// 3. CRITICAL: Renew token on login to prevent Session Fixation attacks
	if err := h.Session.RenewToken(ctx); err != nil {
		slog.Error("Failed to renew session token", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 4. Put user_id into session
	h.Session.Put(ctx, "user_id", user.ID)
	h.Session.Put(ctx, "role", user.Role)

	// 5. Send HX-Redirect header for HTMX clients
	w.Header().Set("HX-Location", "/task")
	w.WriteHeader(http.StatusOK)
}

func (h *authHandler) processLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Destroy the session and clear the cookie
	if err := h.Session.Destroy(ctx); err != nil {
		slog.Error("Failed to destroy session", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// 5. Redirect straight to app dashboard
	w.Header().Set("HX-Redirect", "/auth/login")
	w.WriteHeader(http.StatusOK)
}
