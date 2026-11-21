package main

import (
	"laundry-pos/config"
	"laundry-pos/middleware"
	"laundry-pos/routes"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

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
}

// Handler dipanggil Vercel untuk setiap request.
func Handler(w http.ResponseWriter, r *http.Request) {
	middleware.EnableCORS(routes.InitRoutes()).ServeHTTP(w, r)
}

func main() {
	http.Handle("/", http.HandlerFunc(Handler))
	log.Println("Server berjalan di http://localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
