package services

import (
	"regexp"
	"testing"
	"time"
)

func TestCanTransition(t *testing.T) {
	valid := [][2]string{
		{"pending", "diproses"}, {"pending", "batal"},
		{"diproses", "selesai"}, {"diproses", "batal"},
		{"selesai", "diambil"},
	}
	for _, c := range valid {
		if !CanTransition(c[0], c[1]) {
			t.Fatalf("%s -> %s seharusnya valid", c[0], c[1])
		}
	}
	invalid := [][2]string{
		{"pending", "selesai"},
		{"selesai", "diproses"},
		{"diambil", "diproses"},
		{"batal", "pending"},
		{"pending", "pending"},
		{"", "pending"},
	}
	for _, c := range invalid {
		if CanTransition(c[0], c[1]) {
			t.Fatalf("%s -> %s seharusnya invalid", c[0], c[1])
		}
	}
}

func TestNewInvoiceNumber(t *testing.T) {
	n := NewInvoiceNumber(time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC))
	if !regexp.MustCompile(`^INV-20261121-[A-Z0-9]{4}$`).MatchString(n) {
		t.Fatalf("format nomor nota salah: %s", n)
	}
	// dua kali generate dalam hari yang sama harus unik
	m := NewInvoiceNumber(time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC))
	if m == n {
		t.Fatalf("nomor nota duplikat: %s", n)
	}
}
