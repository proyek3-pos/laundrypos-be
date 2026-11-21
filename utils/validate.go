package utils

import (
	"errors"
	"math"
	"net/mail"
	"regexp"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

func ValidateUsername(s string) error {
	if !usernameRe.MatchString(s) {
		return errors.New("username 3-32 karakter, hanya huruf, angka, dan underscore")
	}
	return nil
}

func ValidatePassword(s string) error {
	if len(s) < 8 {
		return errors.New("password minimal 8 karakter")
	}
	return nil
}

// Email opsional, tapi kalau diisi harus valid.
func ValidateEmail(s string) error {
	if s == "" {
		return nil
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return errors.New("format email tidak valid")
	}
	return nil
}

func ValidateUnit(s string) error {
	if s != "kg" && s != "item" {
		return errors.New("unit harus kg atau item")
	}
	return nil
}

func ValidateName(s string) error {
	if len(s) < 2 || len(s) > 100 {
		return errors.New("nama harus 2-100 karakter")
	}
	return nil
}

var phoneRe = regexp.MustCompile(`^\+?[0-9]{8,15}$`)

func ValidatePhone(s string) error {
	if !phoneRe.MatchString(s) {
		return errors.New("nomor telepon 8-15 digit")
	}
	return nil
}

func ValidatePrice(p float64) error {
	if p <= 0 {
		return errors.New("harga harus lebih dari 0")
	}
	return nil
}

func ValidateQuantity(q float64, unit string) error {
	if q <= 0 {
		return errors.New("jumlah harus lebih dari 0")
	}
	if unit == "item" && q != math.Trunc(q) {
		return errors.New("jumlah untuk unit item harus bilangan bulat")
	}
	return nil
}
