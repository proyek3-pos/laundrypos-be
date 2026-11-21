package services

import (
	"crypto/sha512"
	"fmt"
	"testing"
)

func TestVerifyMidtransSignature(t *testing.T) {
	serverKey := "SB-Mid-server-test"
	orderID, statusCode, gross := "ORD-1", "200", "150000.0"
	want := fmt.Sprintf("%x", sha512.Sum512([]byte(orderID+statusCode+gross+serverKey)))

	if !VerifyMidtransSignature(orderID, statusCode, gross, serverKey, want) {
		t.Fatal("want signature valid")
	}
	if VerifyMidtransSignature(orderID, statusCode, gross, serverKey, "deadbeef") {
		t.Fatal("want signature invalid")
	}
}

func TestMapMidtransStatus(t *testing.T) {
	cases := map[string]string{
		"capture":   "paid",
		"settlement": "paid",
		"pending":   "pending",
		"deny":      "failed",
		"cancel":    "failed",
		"expire":    "failed",
		"failure":   "failed",
	}
	for in, want := range cases {
		if got := MapMidtransStatus(in); got != want {
			t.Fatalf("MapMidtransStatus(%q) = %q, want %q", in, got, want)
		}
	}
	if MapMidtransStatus("apa-aja") != "" {
		t.Fatal("status tak dikenal harus string kosong")
	}
}
