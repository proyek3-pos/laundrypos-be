package main

import (
	"context"
	"fmt"
	"laundry-pos/config"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// Seed membuat/memperbarui satu user admin dari env ADMIN_USERNAME dan
// ADMIN_PASSWORD. Jalankan: go run ./cmd/seed
func main() {
	_ = godotenv.Load()

	if err := config.UseLocalTZ(); err != nil {
		log.Fatal(err)
	}

	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")
	if username == "" || password == "" {
		log.Fatal("ADMIN_USERNAME dan ADMIN_PASSWORD wajib diisi")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err := config.ConnectMongo(cfg); err != nil {
		log.Fatal(err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := config.UserCollection.UpdateOne(ctx,
		bson.M{"username": username},
		bson.M{"$set": bson.M{"username": username, "password": string(hashed), "role": "admin"}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("admin %q siap (upserted: %v)\n", username, res.UpsertedCount)
}
