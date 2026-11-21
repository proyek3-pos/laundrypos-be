package utils

import "testing"

func TestValidateUsername(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"", true},
		{"ab", true},
		{"ada spasi", true},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true}, // 33 karakter
		{"budi_01", false},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false}, // 32 karakter
	}
	for _, c := range cases {
		if err := ValidateUsername(c.in); (err != nil) != c.wantErr {
			t.Fatalf("ValidateUsername(%q) = %v, wantErr %v", c.in, err, c.wantErr)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"", true},
		{"pendek1", true},
		{"password8", false},
	}
	for _, c := range cases {
		if err := ValidatePassword(c.in); (err != nil) != c.wantErr {
			t.Fatalf("ValidatePassword(%q) = %v, wantErr %v", c.in, err, c.wantErr)
		}
	}
}

func TestValidateEmail(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"", false}, // opsional
		{"bukan-email", true},
		{"a@b.co", false},
	}
	for _, c := range cases {
		if err := ValidateEmail(c.in); (err != nil) != c.wantErr {
			t.Fatalf("ValidateEmail(%q) = %v, wantErr %v", c.in, err, c.wantErr)
		}
	}
}

func TestValidateUnit(t *testing.T) {
	for _, u := range []string{"kg", "item"} {
		if err := ValidateUnit(u); err != nil {
			t.Fatalf("ValidateUnit(%q) = %v, want nil", u, err)
		}
	}
	if err := ValidateUnit("pcs"); err == nil {
		t.Fatal("want error untuk unit pcs")
	}
}

func TestValidatePrice(t *testing.T) {
	if err := ValidatePrice(0); err == nil {
		t.Fatal("want error untuk harga 0")
	}
	if err := ValidatePrice(-1); err == nil {
		t.Fatal("want error untuk harga negatif")
	}
	if err := ValidatePrice(15000.5); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

func TestValidateQuantity(t *testing.T) {
	cases := []struct {
		q       float64
		unit    string
		wantErr bool
	}{
		{4.5, "kg", false},
		{0, "kg", true},
		{-1, "kg", true},
		{3, "item", false},
		{4.5, "item", true},
	}
	for _, c := range cases {
		if err := ValidateQuantity(c.q, c.unit); (err != nil) != c.wantErr {
			t.Fatalf("ValidateQuantity(%v, %q) = %v, wantErr %v", c.q, c.unit, err, c.wantErr)
		}
	}
}
