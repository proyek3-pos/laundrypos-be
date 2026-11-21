package handler

import (
	"context"
	"laundry-pos/config"
	"laundry-pos/middleware"
	"laundry-pos/routes"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Inisialisasi sengaja lazy: runtime Vercel baru meng-inject env project
// (VERCEL_ENV_FILE) tepat sebelum Handler dipanggil, bukan saat init package.
var (
	setupOnce sync.Once
	setupErr  error
	router    http.Handler
)

func setup() {
	// .env hanya ada saat lokal (vercel dev); di Vercel env diset lewat dashboard.
	_ = godotenv.Load()

	if err := config.UseLocalTZ(); err != nil {
		setupErr = err
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		setupErr = err
		return
	}
	if err := config.ConnectMongo(cfg); err != nil {
		setupErr = err
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := config.EnsureIndexes(ctx); err != nil {
		setupErr = err
		return
	}

	router = middleware.EnableCORS(middleware.Logger(middleware.Recover(middleware.LimitBody(routes.InitRoutes()))))
}

// Handler dipanggil Vercel untuk setiap request.
func Handler(w http.ResponseWriter, r *http.Request) {
	setupOnce.Do(setup)
	if setupErr != nil {
		log.Printf("setup: %v", setupErr)
		http.Error(w, "Gagal menyiapkan server", http.StatusInternalServerError)
		return
	}
	router.ServeHTTP(w, r)
}
