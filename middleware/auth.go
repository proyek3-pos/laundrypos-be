package middleware

import (
    "laundry-pos/utils"
    "net/http"
)
// EnableCORS menangani header CORS agar frontend dapat mengakses API
func EnableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Mengizinkan CORS dari origin tertentu (frontend Anda)
        w.Header().Add("Access-Control-Allow-Origin", "http://127.0.0.1:5500") // Ganti dengan alamat frontend Anda
        w.Header().Add("Access-Control-Allow-Origin", "https://proyek3-pos.github.io/laundrypos-fe") // Ganti dengan alamat frontend Anda
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Cek jika request adalah preflight request (OPTIONS)
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        // Lanjutkan ke handler berikutnya jika bukan preflight
        next.ServeHTTP(w, r)
    })
}

func RoleMiddleware(role string, next http.Handler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tokenStr := r.Header.Get("Authorization")
        claims, err := utils.ValidateJWT(tokenStr)
        if err != nil || claims.Role != role {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        // Proceed to the next handler if authorized
        next.ServeHTTP(w, r)
    }
}

// func AuthMiddleware(next http.Handler) http.Handler {
//     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         token := r.Header.Get("Authorization")
//         if token == "" {
//             http.Error(w, "Authorization header missing", http.StatusUnauthorized)
//             return
//         }

//         // Token harus diawali dengan "Bearer "
//         if len(token) < 7 || token[:7] != "Bearer " {
//             http.Error(w, "Invalid token format", http.StatusUnauthorized)
//             return
//         }

//         // Ambil token setelah "Bearer "
//         token = token[7:]

//         // Verifikasi token
//         _, err := utils.ValidateJWT(token)
//         if err != nil {
//             http.Error(w, "Invalid token", http.StatusUnauthorized)
//             return
//         }

//         next.ServeHTTP(w, r)
//     })
// }
