package utils

import "testing"

func TestGenerateJWTFailsWithoutSecret(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "")
	_, err := GenerateJWT("u1", "staff", "budi")
	if err == nil {
		t.Fatal("want error saat JWT_SECRET_KEY kosong")
	}
}

func TestGenerateAndValidateJWTRoundtrip(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "testsecret")
	token, err := GenerateJWT("650f1a2b", "admin", "budi")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "650f1a2b" || claims.Role != "admin" || claims.Username != "budi" {
		t.Fatalf("klaim tidak sesuai: %+v", claims)
	}
}

func TestValidateJWTRejectsForeignSignature(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "testsecret")
	token, err := GenerateJWT("u1", "staff", "budi")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("JWT_SECRET_KEY", "rahasia-lain")
	if _, err := ValidateJWT(token); err == nil {
		t.Fatal("want error untuk token dengan secret lain")
	}
}
