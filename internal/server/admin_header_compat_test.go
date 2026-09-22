package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ds2api/internal/auth"
	"github.com/go-chi/chi/v5"
)

func TestAdminAuthorizationHeaderCompat(t *testing.T) {
	cases := []struct{ name, standard, alternate, want string }{
		{"fallback", "", "Bearer test-token", "Bearer test-token"},
		{"standard wins", "Bearer original", "Bearer alternate", "Bearer original"},
		{"invalid standard not bypassed", "Basic original", "Bearer alternate", "Basic original"},
		{"missing", "", "", ""},
		{"reject non bearer", "", "Basic bad", ""},
		{"reject empty token", "", "Bearer ", ""},
		{"trim fallback", "", "  bearer token  ", "bearer token"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/admin/verify", nil)
			if tc.standard != "" {
				req.Header.Set("Authorization", tc.standard)
			}
			if tc.alternate != "" {
				req.Header.Set("X-Ds2api-Admin-Authorization", tc.alternate)
			}
			handler := adminAuthorizationHeaderCompat(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("Authorization"); got != tc.want {
					t.Errorf("unexpected effective authorization")
				}
			}))
			handler.ServeHTTP(httptest.NewRecorder(), req)
			if req.Header.Get("Authorization") != tc.standard {
				t.Fatal("original request mutated")
			}
		})
	}
}

func TestAdminAuthorizationHeaderCompatStillRequiresValidJWT(t *testing.T) {
	t.Setenv("DS2API_ADMIN_KEY", "test-only-random-admin-key-not-for-production")
	t.Setenv("DS2API_JWT_SECRET", "test-only-signing-key")
	token, err := auth.CreateJWT(1)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, token string
		want        int
	}{
		{"valid", token, http.StatusNoContent},
		{"invalid", "not-a-valid-token", http.StatusUnauthorized},
		{"missing", "", http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/admin/verify", nil)
			if tc.token != "" {
				req.Header.Set("X-Ds2api-Admin-Authorization", "Bearer "+tc.token)
			}
			h := adminAuthorizationHeaderCompat(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := auth.VerifyAdminRequest(r); err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Errorf("got %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestAdminAuthorizationHeaderCompatDoesNotApplyToV1(t *testing.T) {
	r := chi.NewRouter()
	check := func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}
	r.Route("/admin", func(ar chi.Router) { ar.Use(adminAuthorizationHeaderCompat); ar.Get("/verify", check) })
	r.Get("/v1/test", check)
	for _, tc := range []struct {
		path string
		want int
	}{{"/admin/verify", 204}, {"/v1/test", 401}} {
		req := httptest.NewRequest("GET", tc.path, nil)
		req.Header.Set("X-Ds2api-Admin-Authorization", "Bearer dummy")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("%s: %d", tc.path, rec.Code)
		}
	}
}
