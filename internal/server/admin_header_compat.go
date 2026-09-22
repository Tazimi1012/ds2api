package server

import (
	"net/http"
	"strings"
)

// adminAuthorizationHeaderCompat is registered only on the /admin router.
// It restores a missing Authorization header; existing admin/JWT checks still
// authenticate the credential. An existing standard header always wins.
func adminAuthorizationHeaderCompat(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSpace(r.Header.Get("Authorization")) == "" {
			value := strings.TrimSpace(r.Header.Get("X-Ds2api-Admin-Authorization"))
			if strings.HasPrefix(strings.ToLower(value), "bearer ") &&
				strings.TrimSpace(value[len("bearer "):]) != "" {
				r = r.Clone(r.Context())
				r.Header.Set("Authorization", value)
			}
		}
		next.ServeHTTP(w, r)
	})
}
