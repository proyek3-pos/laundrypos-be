package controllers

import (
    "context"
    "encoding/json"
    "laundry-pos/config"
    "laundry-pos/models"
    "laundry-pos/utils"
    "net/http"
    "time"

    "golang.org/x/crypto/bcrypt"
    "go.mongodb.org/mongo-driver/bson"
)

// Fungsi untuk registrasi pengguna baru
func Register(w http.ResponseWriter, r *http.Request) {
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Cek apakah username sudah digunakan
    var existingUser models.User
    err := config.UserCollection.FindOne(context.TODO(), bson.M{"username": user.Username}).Decode(&existingUser)
    if err == nil {
        http.Error(w, "Username already exists", http.StatusBadRequest)
        return
    }

    // Hash password sebelum disimpan
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Failed to hash password", http.StatusInternalServerError)
        return
    }
    user.Password = string(hashedPassword)

    // Set role default menjadi "staff" jika role tidak ditentukan
    if user.Role == "" {
        user.Role = "staff"
    }

    // Simpan data pengguna ke database
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err = config.UserCollection.InsertOne(ctx, user)
    if err != nil {
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully", "role": user.Role})
}

func RegisterAdmin(w http.ResponseWriter, r *http.Request) {
    var admin models.User
    if err := json.NewDecoder(r.Body).Decode(&admin); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Validasi API key atau autentikasi
    apiKey := r.Header.Get("X-API-KEY")
    if apiKey != "your-secret-api-key" { // Ganti dengan API key Anda
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Cek apakah username sudah digunakan
    var existingAdmin models.User
    err := config.UserCollection.FindOne(context.TODO(), bson.M{"username": admin.Username}).Decode(&existingAdmin)
    if err == nil {
        http.Error(w, "Username already exists", http.StatusBadRequest)
        return
    }

    // Hash password sebelum disimpan
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Failed to hash password", http.StatusInternalServerError)
        return
    }
    admin.Password = string(hashedPassword)

    // Set role menjadi "admin"
    admin.Role = "admin"

    // Simpan data admin ke database
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err = config.UserCollection.InsertOne(ctx, admin)
    if err != nil {
        http.Error(w, "Failed to create admin", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Admin registered successfully", "role": admin.Role})
}


// Fungsi untuk login
func Login(w http.ResponseWriter, r *http.Request) {
    var creds struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Find user by username
    var user models.User
    err := config.UserCollection.FindOne(context.TODO(), bson.M{"username": creds.Username}).Decode(&user)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Verify password
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Generate JWT token with userID and role
    token, err := utils.GenerateJWT(user.ID, user.Role) // Pass user role here
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
    }

    // Send the token back in the response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"token": token})
}
