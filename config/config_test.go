package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestUseLocalTZ(t *testing.T) {
	t.Setenv("APP_TZ", "")
	if err := UseLocalTZ(); err != nil {
		t.Fatalf("Asia/Jakarta harus tersedia: %v", err)
	}
	t1 := time.Date(2026, 11, 21, 0, 0, 0, 0, time.Local)
	if t1.UTC().Hour() != 17 {
		t.Fatalf("default harus Asia/Jakarta (UTC+7), got %v", t1.UTC())
	}

	t.Setenv("APP_TZ", "Bukan/Timezone")
	if err := UseLocalTZ(); err == nil {
		t.Fatal("timezone tidak dikenal harus error")
	}
}

func TestLoadConfigFailsWithoutMongoURI(t *testing.T) {
	t.Setenv("MONGO_URI", "")
	t.Setenv("JWT_SECRET_KEY", "secret")
	_, err := LoadConfig()
	if err == nil || !strings.Contains(err.Error(), "MONGO_URI") {
		t.Fatalf("want MONGO_URI error, got %v", err)
	}
}

func TestLoadConfigFailsWithoutJWTSecret(t *testing.T) {
	t.Setenv("MONGO_URI", "mongodb://example")
	t.Setenv("JWT_SECRET_KEY", "")
	_, err := LoadConfig()
	if err == nil || !strings.Contains(err.Error(), "JWT_SECRET_KEY") {
		t.Fatalf("want JWT_SECRET_KEY error, got %v", err)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("MONGO_URI", "mongodb://example")
	t.Setenv("JWT_SECRET_KEY", "secret")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DBName != "laundry-pos" {
		t.Fatalf("default DBName = %q, want laundry-pos", cfg.DBName)
	}
	if len(cfg.CORSOrigins) == 0 {
		t.Fatal("CORSOrigins should have a default, not be empty")
	}
}

func TestParseCORSOrigins(t *testing.T) {
	got := ParseCORSOrigins(" http://a:1 , http://b:2 ,,")
	want := []string{"http://a:1", "http://b:2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseCORSOrigins = %v, want %v", got, want)
	}
}
