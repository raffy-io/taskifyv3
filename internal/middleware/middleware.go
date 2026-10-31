package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
)

// 1. Define a custom type for context keys to avoid collisions
type contextKey string

// 2. Declare the constant used to store and retrieve the user ID in context
const UserIDKey contextKey = "user_id"

// 3. Helper function to extract the User ID from context anywhere in your app
func GetUserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(UserIDKey).(int64)
	return id, ok
}


// RequireAuth checks the SCS session for a user_id and handles HTMX requests gracefully
func RequireAuth(sessionManager *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 1. Check if the user_id exists in the SCS session
			if !sessionManager.Exists(ctx, "user_id") {
		
				// If it's an HTMX request, send HX-Redirect header so the browser 
				// performs a full page navigation instead of swapping the login HTML into a target div.
				if r.Header.Get("HX-Request") == "true" {
					slog.Info("Unauthenticated HTMX request, redirecting to login", "path", r.URL.Path)
					w.Header().Set("HX-Redirect", "/auth/login")
					w.WriteHeader(http.StatusSeeOther)
					return
				}

				slog.Info("Unauthenticated request, redirecting to login", "path", r.URL.Path)
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			// 2. Fetch the user ID from the session
			userID := sessionManager.GetInt64(ctx, "user_id")

			// 3. Store the user ID in the request context using your custom type key
			ctx = context.WithValue(ctx, UserIDKey, userID)

			// 4. Move on to the actual handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// NoSurf adds CSRF protection to the router
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad or missing CSRF token", http.StatusForbidden)

		slog.Error("bad or missing CSRF token",
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
		)
	}))

	return csrfHandler
}
