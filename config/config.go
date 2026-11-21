package config

import (
	"errors"
	"os"
	"strings"
	"time"

	// Embed database timezone supaya APP_TZ jalan di semua platform.
	_ "time/tzdata"
)

type Config struct {
	MongoURI          string
	DBName            string
	JWTSecret         string
	MidtransServerKey string
	MidtransClientKey string
	CORSOrigins       []string
}

var DefaultCORSOrigins = []string{
	"http://127.0.0.1:5500",
	"http://localhost:5173",
	"https://proyek3-pos.github.io/laundrypos-fe",
}

func LoadConfig() (Config, error) {
	cfg := Config{
		MongoURI:          os.Getenv("MONGO_URI"),
		DBName:            os.Getenv("DB_NAME"),
		JWTSecret:         os.Getenv("JWT_SECRET_KEY"),
		MidtransServerKey: os.Getenv("MIDTRANS_SERVER_KEY"),
		MidtransClientKey: os.Getenv("MIDTRANS_CLIENT_KEY"),
		CORSOrigins:       ParseCORSOrigins(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}
	if cfg.DBName == "" {
		cfg.DBName = "laundry-pos"
	}
	if len(cfg.CORSOrigins) == 0 {
		cfg.CORSOrigins = DefaultCORSOrigins
	}
	if cfg.MongoURI == "" {
		return cfg, errors.New("env MONGO_URI wajib diisi")
	}
	if cfg.JWTSecret == "" {
		return cfg, errors.New("env JWT_SECRET_KEY wajib diisi")
	}
	return cfg, nil
}

func ParseCORSOrigins(s string) []string {
	var out []string
	for _, o := range strings.Split(s, ",") {
		if o = strings.TrimSpace(o); o != "" {
			out = append(out, o)
		}
	}
	return out
}

// UseLocalTZ menyetel time.Local dari APP_TZ; default Asia/Jakarta supaya
// batas "hari ini" di laporan mengikuti waktu toko, bukan UTC server.
func UseLocalTZ() error {
	name := os.Getenv("APP_TZ")
	if name == "" {
		name = "Asia/Jakarta"
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return err
	}
	time.Local = loc
	return nil
}
