package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"laundry-pos/utils"
)

func reqWithClaims(c utils.JWTClaims) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	return r.WithContext(contextWithClaims(r.Context(), c))
}

func TestEnableCORSAllowsListedOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://a.test")
	h := EnableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://a.test")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://a.test" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want http://a.test", got)
	}
}

func TestEnableCORSRejectsUnknownOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://a.test")
	h := EnableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://evil.test")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("origin tak dikenal tidak boleh diizinkan, got %q", got)
	}
}

func TestEnableCORSPreflight(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://a.test")
	h := EnableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Origin", "http://a.test")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("preflight = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://a.test" {
		t.Fatalf("preflight ACAO = %q", got)
	}
}

func TestAuthMiddlewareSetsClaims(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "testsecret")
	token, err := utils.GenerateJWT("uid1", "admin", "budi")
	if err != nil {
		t.Fatal(err)
	}

	var got utils.JWTClaims
	h := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = ClaimsFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got.Role != "admin" || got.UserID != "uid1" {
		t.Fatalf("claims dari context = %+v", got)
	}
}

func TestAuthMiddlewareRejectsBadToken(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "testsecret")
	h := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer bocor")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("token palsu = %d, want 401", rec.Code)
	}
}

func TestRequireRole(t *testing.T) {
	admin := utils.JWTClaims{Role: "admin"}
	staff := utils.JWTClaims{Role: "staff"}
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	rec := httptest.NewRecorder()
	RequireRole("admin", ok).ServeHTTP(rec, reqWithClaims(staff))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("staff ke endpoint admin = %d, want 403", rec.Code)
	}

	rec = httptest.NewRecorder()
	RequireRole("admin", ok).ServeHTTP(rec, reqWithClaims(admin))
	if rec.Code != http.StatusOK {
		t.Fatalf("admin ke endpoint admin = %d, want 200", rec.Code)
	}

	rec = httptest.NewRecorder()
	RequireRole("admin", ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("tanpa claims = %d, want 401", rec.Code)
	}
}
