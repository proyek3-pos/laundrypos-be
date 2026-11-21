package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

var transitions = map[string][]string{
	"pending":  {"diproses", "batal"},
	"diproses": {"selesai", "batal"},
	"selesai":  {"diambil"},
	"diambil":  {},
	"batal":    {},
}

func CanTransition(from, to string) bool {
	for _, next := range transitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

const invoiceCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// NewInvoiceNumber menghasilkan INV-YYYYMMDD-XXXX dengan suffix acak.
// Karakter ambigu (I, O, 0, 1) sengaja dibuang supaya mudah dibaca lisan.
func NewInvoiceNumber(now time.Time) string {
	suffix := make([]byte, 4)
	for i := range suffix {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(invoiceCharset))))
		suffix[i] = invoiceCharset[n.Int64()]
	}
	return fmt.Sprintf("INV-%s-%s", now.Format("20060102"), suffix)
}
