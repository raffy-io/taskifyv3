package middleware

import (
	"net/http"
	"strings"
)

// UserAgentValidator checks if the User-Agent header is present and valid.
func UserAgentValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")

		// Example: Block empty user agents (common for basic bots/scrapers)
		if strings.TrimSpace(ua) == "" {
			http.Error(w, "Forbidden: Missing User-Agent", http.StatusForbidden)
			return
		}

		// You can also inject the user agent into request context if needed down the chain:
		// ctx := context.WithValue(r.Context(), "user_agent", ua)
		// next.ServeHTTP(w, r.WithContext(ctx))

		next.ServeHTTP(w, r)
	})
}
