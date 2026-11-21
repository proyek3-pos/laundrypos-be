package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	MongoURI          string
	DBName            string
	JWTSecret         string
	MidtransServerKey string
	MidtransClientKey string
	CORSOrigins       []string
}

var defaultCORSOrigins = []string{
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
		CORSOrigins:       parseCORSOrigins(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}
	if cfg.DBName == "" {
		cfg.DBName = "laundry-pos"
	}
	if len(cfg.CORSOrigins) == 0 {
		cfg.CORSOrigins = defaultCORSOrigins
	}
	if cfg.MongoURI == "" {
		return cfg, errors.New("env MONGO_URI wajib diisi")
	}
	if cfg.JWTSecret == "" {
		return cfg, errors.New("env JWT_SECRET_KEY wajib diisi")
	}
	return cfg, nil
}

func parseCORSOrigins(s string) []string {
	var out []string
	for _, o := range strings.Split(s, ",") {
		if o = strings.TrimSpace(o); o != "" {
			out = append(out, o)
		}
	}
	return out
}
