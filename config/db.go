package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/joho/godotenv" // Import godotenv untuk memuat file .env
)

// DB adalah variabel global untuk koneksi database
var DB *gorm.DB

// InitSupabase untuk menginisialisasi koneksi ke Supabase
func InitDB() error {
	// Memuat file .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
		return err
	}

	// Mengambil nilai variabel dari .env
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbSslMode := os.Getenv("DB_SSLMODE")

	// Membuat connection string PostgreSQL
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSslMode)

	// Membuka koneksi ke database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return err
	}

	log.Println("Connected to Supabase database successfully!")
	DB = db
	return nil
}
