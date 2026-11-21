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

// MapMidtransStatus memetakan status mentah Midtrans ke status internal.
// Status tak dikenal mengembalikan string kosong supaya caller bisa melewatinya.
func MapMidtransStatus(transactionStatus string) string {
	switch transactionStatus {
	case "capture", "settlement":
		return "paid"
	case "pending":
		return "pending"
	case "deny", "cancel", "expire", "failure":
		return "failed"
	default:
		return ""
	}
}
