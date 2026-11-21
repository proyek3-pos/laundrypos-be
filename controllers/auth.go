package controllers

import (
	"context"
	"encoding/json"
	"laundry-pos/config"
	"laundry-pos/models"
	"laundry-pos/utils"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}
	if err := utils.ValidateUsername(user.Username); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.ValidatePassword(user.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Gagal mengenkripsi password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashed)
	user.Role = "staff"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := config.UserCollection.InsertOne(ctx, user); err != nil {
		// Index unik username yang menahan race antar dua register bersamaan.
		if mongo.IsDuplicateKeyError(err) {
			http.Error(w, "Username sudah dipakai", http.StatusConflict)
			return
		}
		http.Error(w, "Gagal membuat user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Registrasi berhasil"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}

	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := config.UserCollection.FindOne(ctx, bson.M{"username": creds.Username}).Decode(&user)
	if err != nil {
		http.Error(w, "Username atau password salah", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		http.Error(w, "Username atau password salah", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateJWT(user.ID.Hex(), user.Role, user.Username)
	if err != nil {
		http.Error(w, "Gagal membuat token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":    token,
		"role":     user.Role,
		"username": user.Username,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
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
	if claims.ExpiresAt == nil {
		http.Error(w, "Token tidak valid", http.StatusUnauthorized)
		return
	}

	// Token dicabut sampai kedaluwarsa; TTL index Mongo yang menghapus dokumennya.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = config.RevokedTokenCollection.InsertOne(ctx, bson.M{
		"token":     token[7:],
		"expiresAt": claims.ExpiresAt.Time,
	})
	if err != nil {
		http.Error(w, "Gagal logout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout berhasil"})
}

// Health dipakai monitoring dan smoke test.
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
