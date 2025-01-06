package utils

import (
    "github.com/dgrijalva/jwt-go"
    "time"
    "os"
    "sync"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY")) // Load the secret key from environment variables

// JWTClaims defines the structure of the JWT claims
type JWTClaims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"` // Add role to the claims
    jwt.StandardClaims
}

// GenerateJWT generates a JWT token for the user with ID and role
func GenerateJWT(userID, role, username string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour) // Mengatur token berlaku 24 jam
    claims := &JWTClaims{
        UserID: userID,
        Role:   role,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(), // Waktu kedaluwarsa
            IssuedAt:  time.Now().Unix(),     // Waktu diterbitkan
        },
    }

    // Membuat token baru dengan klaim
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // Tanda tangan token dengan secret key
    return token.SignedString(jwtKey)
}


// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenStr string) (*JWTClaims, error) {
    claims := &JWTClaims{}

    // Periksa apakah token ada di blacklist
    if IsTokenBlacklisted(tokenStr) {
        return nil, jwt.ErrSignatureInvalid // Token sudah logout
    }

    // Parse token dan klaimnya
    token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })

    if err != nil {
        // Token tidak valid
        if err == jwt.ErrSignatureInvalid {
            return nil, err
        }
        return nil, jwt.NewValidationError("Invalid token format", jwt.ValidationErrorMalformed)
    }

    if !token.Valid {
        return nil, jwt.NewValidationError("Invalid token", jwt.ValidationErrorSignatureInvalid)
    }

    // Kembalikan klaim jika valid
    return claims, nil
}


var blacklist = struct {
	sync.RWMutex
	tokens map[string]time.Time
}{tokens: make(map[string]time.Time)}

// AddToBlacklist menambahkan token ke daftar blacklist
func AddToBlacklist(token string, expiry time.Time) {
    blacklist.Lock()
    defer blacklist.Unlock()
    blacklist.tokens[token] = expiry

    // Membersihkan token kedaluwarsa dari blacklist
    for t, exp := range blacklist.tokens {
        if time.Now().After(exp) {
            delete(blacklist.tokens, t) // Hapus token yang sudah expired
        }
    }
}

func IsTokenBlacklisted(token string) bool {
    blacklist.RLock()
    defer blacklist.RUnlock()
    expiry, exists := blacklist.tokens[token]
    if !exists {
        return false
    }

    // Jika token sudah expired, hapus dari blacklist
    if time.Now().After(expiry) {
        delete(blacklist.tokens, token) // Hapus saat dicek
        return false
    }
    return true
}
