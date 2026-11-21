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
