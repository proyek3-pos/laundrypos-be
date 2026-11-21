package main

import (
	"context"
	"laundry-pos/config"
	"laundry-pos/middleware"
	"laundry-pos/routes"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

var router http.Handler

func init() {
	// .env hanya ada saat lokal; di Vercel env diset lewat dashboard.
	_ = godotenv.Load()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := config.ConnectMongo(cfg); err != nil {
		log.Fatalf("mongo: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := config.EnsureIndexes(ctx); err != nil {
		log.Fatalf("indexes: %v", err)
	}

	router = middleware.EnableCORS(middleware.Logger(middleware.Recover(middleware.LimitBody(routes.InitRoutes()))))
}

// Handler dipanggil Vercel untuk setiap request.
func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}

func main() {
	http.Handle("/", http.HandlerFunc(Handler))
	log.Println("Server berjalan di http://localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
