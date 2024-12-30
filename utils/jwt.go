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
func GenerateJWT(userID, role string) (string, error) {
    expirationTime := time.Now().Add(1 * time.Hour)
    claims := &JWTClaims{
        UserID: userID,
        Role:   role,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    // Create a new JWT token with the claims and the signing method
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // Sign the token with the secret key
    return token.SignedString(jwtKey)
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenStr string) (*JWTClaims, error) {
    claims := &JWTClaims{}
    
    // Parse the token and extract the claims
    token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })
    
    if err != nil || !token.Valid {
        return nil, err // Return nil if invalid token
    }

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
}

// IsTokenBlacklisted mengecek apakah token ada dalam blacklist
func IsTokenBlacklisted(token string) bool {
	blacklist.RLock()
	defer blacklist.RUnlock()
	expiry, exists := blacklist.tokens[token]
	if !exists {
		return false
	}
	return time.Now().Before(expiry)
}
