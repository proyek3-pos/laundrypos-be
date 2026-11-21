package services

import (
	"crypto/sha512"
	"encoding/hex"
	"os"
	"strings"

	"github.com/veritrans/go-midtrans"
)

// shortcut: masih Sandbox, butuh MIDTRANS_ENV saat ganti ke produksi.
func MidtransClient() *midtrans.Client {
	c := midtrans.NewClient()
	c.ServerKey = os.Getenv("MIDTRANS_SERVER_KEY")
	c.ClientKey = os.Getenv("MIDTRANS_CLIENT_KEY")
	c.APIEnvType = midtrans.Sandbox
	return &c
}

// Signature Midtrans: SHA512(order_id + status_code + gross_amount + serverKey).
func VerifyMidtransSignature(orderID, statusCode, grossAmount, serverKey, signature string) bool {
	sum := sha512.Sum512([]byte(orderID + statusCode + grossAmount + serverKey))
	want := hex.EncodeToString(sum[:])
	return strings.EqualFold(want, signature)
}
