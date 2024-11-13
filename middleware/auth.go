package middleware

import (
    "laundry-pos/utils"
    "net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Authorization header missing", http.StatusUnauthorized)
            return
        }

        // Token harus diawali dengan "Bearer "
        if len(token) < 7 || token[:7] != "Bearer " {
            http.Error(w, "Invalid token format", http.StatusUnauthorized)
            return
        }

        // Ambil token setelah "Bearer "
        token = token[7:]

        // Verifikasi token
        _, err := utils.ValidateJWT(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}
