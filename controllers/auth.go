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
		utils.Err(w, http.StatusBadRequest, "Input tidak valid")
		return
	}
	if err := utils.ValidateUsername(user.Username); err != nil {
		utils.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := utils.ValidatePassword(user.Password); err != nil {
		utils.Err(w, http.StatusBadRequest, err.Error())
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal mengenkripsi password")
		return
	}
	user.Password = string(hashed)
	user.Role = "staff"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := config.UserCollection.InsertOne(ctx, user); err != nil {
		// Index unik username yang menahan race antar dua register bersamaan.
		if mongo.IsDuplicateKeyError(err) {
			utils.Err(w, http.StatusConflict, "Username sudah dipakai")
			return
		}
		utils.Err(w, http.StatusInternalServerError, "Gagal membuat user")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]string{"message": "Registrasi berhasil"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		utils.Err(w, http.StatusBadRequest, "Input tidak valid")
		return
	}

	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := config.UserCollection.FindOne(ctx, bson.M{"username": creds.Username}).Decode(&user)
	if err != nil {
		utils.Err(w, http.StatusUnauthorized, "Username atau password salah")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		utils.Err(w, http.StatusUnauthorized, "Username atau password salah")
		return
	}

	token, err := utils.GenerateJWT(user.ID.Hex(), user.Role, user.Username)
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal membuat token")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"token":    token,
		"role":     user.Role,
		"username": user.Username,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if len(token) < 8 || token[:7] != "Bearer " {
		utils.Err(w, http.StatusUnauthorized, "Token tidak valid")
		return
	}
	claims, err := utils.ValidateJWT(token[7:])
	if err != nil {
		utils.Err(w, http.StatusUnauthorized, "Token tidak valid")
		return
	}
	if claims.ExpiresAt == nil {
		utils.Err(w, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	// Token dicabut sampai kedaluwarsa; TTL index Mongo yang menghapus dokumennya.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := config.RevokedTokenCollection.InsertOne(ctx, bson.M{
		"token":     token[7:],
		"expiresAt": claims.ExpiresAt.Time,
	}); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal logout")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Logout berhasil"})
}

// Health dipakai monitoring dan smoke test.
func Health(w http.ResponseWriter, r *http.Request) {
	utils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
