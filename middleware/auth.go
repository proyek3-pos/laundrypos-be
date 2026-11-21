package middleware

import (
	"context"
	"net/http"
	"os"

	"laundry-pos/config"
	"laundry-pos/utils"
)

func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := map[string]bool{}
		for _, o := range corsOrigins() {
			allowed[o] = true
		}
		if origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func corsOrigins() []string {
	if s := os.Getenv("CORS_ALLOWED_ORIGINS"); s != "" {
		return config.ParseCORSOrigins(s)
	}
	return config.DefaultCORSOrigins
}

type ctxKey struct{}

func contextWithClaims(ctx context.Context, claims utils.JWTClaims) context.Context {
	return context.WithValue(ctx, ctxKey{}, claims)
}

func ClaimsFromContext(ctx context.Context) (utils.JWTClaims, bool) {
	claims, ok := ctx.Value(ctxKey{}).(utils.JWTClaims)
	return claims, ok
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if len(token) < 8 || token[:7] != "Bearer " {
			http.Error(w, "Token tidak valid", http.StatusUnauthorized)
			return
		}
		claims, err := utils.ValidateJWT(token[7:])
		if err != nil {
			http.Error(w, "Token tidak valid", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(contextWithClaims(r.Context(), *claims)))
	})
}

func RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "Belum login", http.StatusUnauthorized)
			return
		}
		if claims.Role != role {
			http.Error(w, "Akses ditolak", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
