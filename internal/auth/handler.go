package auth

import (
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/raffy-io/taskifyv3/internal/db"
	"github.com/raffy-io/taskifyv3/internal/utils"
)

type authHandler struct {
	Queries *db.Queries
	Session *scs.SessionManager
}

func (h *authHandler) processRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w,"Bad request",http.StatusBadRequest)
		return
	}

	form := &authForm{
		Username: r.PostFormValue("username"),
		Email: r.PostFormValue("email"),
		PasswordHash: r.PostFormValue("password"),
	}

	if !form.validate() {
		slog.Error("Failed to register", "err", form.Errors)
		http.Error(w, "Validation failed", http.StatusBadRequest)
		return
	}

	hashedPassword,err := utils.HashPassword(form.PasswordHash)
	if err != nil {
		http.Error(w,"Invalid Password",http.StatusInternalServerError)
		return
	}

	user, err := h.Queries.CreateUser(ctx, db.CreateUserParams{
		Username: form.Username,
		Email: form.Email,
		PasswordHash: hashedPassword,
	})

	if err != nil {
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }

	slog.Info("Succesfully registered", "user", user.Username)
}
func (h *authHandler) processLogin(w http.ResponseWriter, r *http.Request)    {
	ctx := r.Context()

	// 1. Fetch user & validate password...
    user, err := h.Queries.GetUserByEmail(ctx, r.PostFormValue("email"))
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

	if !utils.CheckPassword(r.PostFormValue("password"), user.PasswordHash) {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

	// 2. CRITICAL: Renew token on login to prevent Session Fixation attacks
    if err := h.Session.RenewToken(ctx); err != nil {
        slog.Error("Failed to renew session token", "err", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

	// 3. Put user_id into session
    h.Session.Put(ctx, "user_id", user.ID)
    slog.Info("User logged in successfully", "user", user.Username)
}

func (h *authHandler) processLogout(w http.ResponseWriter, r *http.Request)   {
	ctx := r.Context()

	// 1. Destroy the session and clear the cookie
    if err := h.Session.Destroy(ctx); err != nil {
        slog.Error("Failed to destroy session", "err", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    slog.Info("User logged out successfully")
}
