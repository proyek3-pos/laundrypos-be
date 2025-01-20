package middleware

import (
    "laundry-pos/utils"
    "net/http"
)
// EnableCORS menangani header CORS agar frontend dapat mengakses API
func EnableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        
        // Daftar origin yang diperbolehkan
        allowedOrigins := map[string]bool{
            "http://127.0.0.1:5500":                       true,
            "https://proyek3-pos.github.io/laundrypos-fe": true,
            "https://proyek3-pos.github.io/swagger":       true,
            "https://proyek3-pos.github.io":               true,
        }

        // Periksa apakah origin diizinkan
        if allowedOrigins[origin] {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Credentials", "true") // Mengizinkan kredensial
        } else {
            // If the origin is not allowed, set CORS headers to prevent access
            w.Header().Set("Access-Control-Allow-Origin", "null")
        }

        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Tangani preflight request (OPTIONS)
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}





func RoleMiddleware(requiredRole string, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Ambil token dari header Authorization
        token := r.Header.Get("Authorization")
        if len(token) < 7 || token[:7] != "Bearer " {
            http.Error(w, "Unauthorized: Token Format Salah", http.StatusUnauthorized)
            return
        }

        token = token[7:]

        // Validasi token JWT
        claims, err := utils.ValidateJWT(token)
        if err != nil {
            http.Error(w, "Token tidak valid atau kedaluwarsa", http.StatusUnauthorized)
            return
        }

        // Periksa apakah role valid
        allowedRoles := map[string]bool{
            "admin": true,
            "staff": true,
        }

        if !allowedRoles[claims.Role] {
            http.Error(w, "Unknown role. Please contact administrator.", http.StatusForbidden)
            return
        }

        // Periksa apakah role cocok
        if claims.Role != requiredRole {
            http.Error(w, "Access denied", http.StatusForbidden)
            return
        }

        // Lanjutkan ke handler yang dimaksud
        next.ServeHTTP(w, r)
    })
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


func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Ambil token dari header Authorization
        token := r.Header.Get("Authorization")
        if len(token) < 7 || token[:7] != "Bearer " {
            http.Error(w, "Unauthorized: Token Tidak Ditemukan", http.StatusUnauthorized)
            return
        }

        // Ambil token tanpa 'Bearer '
        token = token[7:]

        // Periksa apakah token ada dalam blacklist
        if utils.IsTokenBlacklisted(token) {
            http.Error(w, "Token telah di-logout", http.StatusUnauthorized)
            return
        }

        // Validasi token JWT
        claims, err := utils.ValidateJWT(token)
        if err != nil {
            http.Error(w, "Token tidak valid atau kedaluwarsa", http.StatusUnauthorized)
            return
        }

        // Tambahkan informasi pengguna ke request untuk digunakan oleh handler selanjutnya
        r.Header.Set("X-User-ID", claims.UserID)
        r.Header.Set("X-Role", claims.Role)

        // Lanjutkan ke handler berikutnya
        next.ServeHTTP(w, r)
    })
}


func RoleAuthorization(requiredRole string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Authorization token required", http.StatusUnauthorized)
            return
        }

        claims, err := utils.ValidateJWT(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Validasi role, tambahkan logika untuk menangani role yang tidak dikenal
        if claims.Role != "admin" && claims.Role != "staff" {
            http.Error(w, "Unknown role. Please contact administrator.", http.StatusForbidden)
            return
        }

        if claims.Role != requiredRole {
            http.Error(w, "Access denied", http.StatusForbidden)
            return
        }

        // Lanjut ke handler jika role cocok
        http.DefaultServeMux.ServeHTTP(w, r)
    }
}