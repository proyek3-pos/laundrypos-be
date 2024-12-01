package controllers

import (
	"encoding/json"
	"laundry-pos/config"
	"laundry-pos/models"
	"laundry-pos/utils"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	err := config.DB.Where("username = ?", user.Username).First(&existingUser).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		http.Error(w, "Error checking username", http.StatusInternalServerError)
		return
	}
	if existingUser.ID != uuid.Nil {
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

	// Simpan data pengguna ke database Supabase
	if err := config.DB.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
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

	// Mencari pengguna berdasarkan username
	var user models.User
	err := config.DB.Where("username = ?", creds.Username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Error fetching user", http.StatusInternalServerError)
		return
	}

	// Memverifikasi password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token JWT jika login berhasil
	token, err := utils.GenerateJWT(user.ID.String(), user.Role) // Pass user.ID and user.Role
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
