package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"laundry-pos/config"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// Test alur end-to-end handler. Butuh Mongo via TEST_MONGO_URI; database
// laundry-pos-test dibersihkan setiap jalan. Tanpa env, semua test ini skip.
func setupTest(t *testing.T) {
	t.Helper()
	if os.Getenv("TEST_MONGO_URI") == "" {
		t.Skip("TEST_MONGO_URI tidak diset")
	}
	t.Setenv("JWT_SECRET_KEY", "testsecret")
	cfg := config.Config{
		MongoURI: os.Getenv("TEST_MONGO_URI"),
		DBName:   "laundry-pos-test",
	}
	if err := config.ConnectMongo(cfg); err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, name := range []string{"user", "customers", "service", "transactions", "payments", "revoked_tokens"} {
		if err := config.Client.Database(cfg.DBName).Collection(name).Drop(ctx); err != nil {
			t.Fatalf("drop %s: %v", name, err)
		}
	}
	if err := config.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
}

func request(t *testing.T, method, path, token, body string) (int, map[string]any) {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	InitRoutes().ServeHTTP(rec, req)

	var out map[string]any
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("response bukan JSON: %s", rec.Body.String())
		}
	}
	return rec.Code, out
}

func loginAs(t *testing.T, username, password string) string {
	t.Helper()
	code, body := request(t, "POST", "/login", "", fmt.Sprintf(`{"username":%q,"password":%q}`, username, password))
	if code != http.StatusOK {
		t.Fatalf("login %s: %d %v", username, code, body)
	}
	data := body["data"].(map[string]any)
	return data["token"].(string)
}

func TestFullFlow(t *testing.T) {
	setupTest(t)

	// register staff + login
	code, body := request(t, "POST", "/register", "", `{"username":"staff1","password":"password123"}`)
	if code != http.StatusCreated {
		t.Fatalf("register: %d %v", code, body)
	}
	code, body = request(t, "POST", "/register", "", `{"username":"staff1","password":"password123"}`)
	if code != http.StatusConflict {
		t.Fatalf("register duplikat: %d %v, want 409", code, body)
	}
	staffToken := loginAs(t, "staff1", "password123")

	// admin disuntik langsung lalu login
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if _, err := config.UserCollection.InsertOne(context.Background(), map[string]any{
		"username": "admin1", "password": string(hashed), "role": "admin",
	}); err != nil {
		t.Fatal(err)
	}
	adminToken := loginAs(t, "admin1", "password123")

	// endpoint tanpa token ditolak
	if code, _ := request(t, "GET", "/customers", "", ""); code != http.StatusUnauthorized {
		t.Fatalf("tanpa token: %d, want 401", code)
	}

	// customer
	code, body = request(t, "POST", "/customers", staffToken,
		`{"fullName":"Budi Santoso","phoneNumber":"081234567890","email":"budi@contoh.id"}`)
	if code != http.StatusCreated {
		t.Fatalf("create customer: %d %v", code, body)
	}
	customerID := body["data"].(map[string]any)["id"].(string)

	if code, _ = request(t, "POST", "/customers", staffToken, `{"fullName":"X","phoneNumber":"salah","email":""}`); code != http.StatusBadRequest {
		t.Fatalf("customer tidak valid: %d, want 400", code)
	}

	// layanan: staff dilarang, admin boleh
	if code, _ = request(t, "POST", "/services", staffToken, `{"serviceName":"Cuci Kering","unitPrice":7000,"unit":"kg"}`); code != http.StatusForbidden {
		t.Fatalf("create service oleh staff: %d, want 403", code)
	}
	code, body = request(t, "POST", "/services", adminToken, `{"serviceName":"Cuci Kering","unitPrice":7000,"unit":"kg"}`)
	if code != http.StatusCreated {
		t.Fatalf("create service: %d %v", code, body)
	}
	serviceID := body["data"].(map[string]any)["id"].(string)

	// transaksi: kg desimal diterima, total dihitung ulang server
	code, body = request(t, "POST", "/transactions", staffToken, fmt.Sprintf(
		`{"customerId":%q,"items":[{"serviceId":%q,"quantity":4.5}]}`, customerID, serviceID))
	if code != http.StatusCreated {
		t.Fatalf("create transaction: %d %v", code, body)
	}
	tx := body["data"].(map[string]any)
	if tx["totalAmount"] != 31500.0 {
		t.Fatalf("total = %v, want 31500", tx["totalAmount"])
	}
	if !regexp.MustCompile(`^INV-\d{8}-[A-Z0-9]{4}$`).MatchString(tx["invoiceNumber"].(string)) {
		t.Fatalf("invoiceNumber = %v", tx["invoiceNumber"])
	}
	txID := tx["id"].(string)

	// patch status: lompatan ditolak, urutan benar diterima
	if code, _ = request(t, "PATCH", "/transactions/"+txID+"/status", staffToken, `{"status":"selesai"}`); code != http.StatusBadRequest {
		t.Fatalf("lompat status: %d, want 400", code)
	}
	if code, _ = request(t, "PATCH", "/transactions/"+txID+"/status", staffToken, `{"status":"diproses"}`); code != http.StatusOK {
		t.Fatalf("status diproses: %d", code)
	}

	// edit item setelah bukan pending ditolak
	if code, _ = request(t, "PUT", "/transactions/"+txID, staffToken, fmt.Sprintf(
		`{"customerId":%q,"items":[{"serviceId":%q,"quantity":1}]}`, customerID, serviceID)); code != http.StatusBadRequest {
		t.Fatalf("edit setelah diproses: %d, want 400", code)
	}

	// cash payment: kurang ditolak, pas menghitung kembalian, dobel ditolak
	if code, _ = request(t, "POST", "/payments/cash", staffToken, fmt.Sprintf(
		`{"transactionId":%q,"amountPaid":30000}`, txID)); code != http.StatusBadRequest {
		t.Fatalf("bayar kurang: %d, want 400", code)
	}
	code, body = request(t, "POST", "/payments/cash", staffToken, fmt.Sprintf(
		`{"transactionId":%q,"amountPaid":36500}`, txID))
	if code != http.StatusCreated {
		t.Fatalf("cash payment: %d %v", code, body)
	}
	if body["data"].(map[string]any)["change"] != 5000.0 {
		t.Fatalf("kembalian = %v, want 5000", body["data"].(map[string]any)["change"])
	}
	if code, _ = request(t, "POST", "/payments/cash", staffToken, fmt.Sprintf(
		`{"transactionId":%q,"amountPaid":36500}`, txID)); code != http.StatusConflict {
		t.Fatalf("dobel bayar: %d, want 409", code)
	}

	// midtrans pada transaksi lunas ditolak
	if code, _ = request(t, "POST", "/payments/midtrans", staffToken, fmt.Sprintf(`{"transactionId":%q}`, txID)); code != http.StatusConflict {
		t.Fatalf("midtrans sudah lunas: %d, want 409", code)
	}

	// webhook dengan signature palsu ditolak
	if code, _ = request(t, "POST", "/webhook/midtrans", "", `{"order_id":"palsu","status_code":"200","gross_amount":"31500.00","signature_key":"palsu","transaction_status":"settlement"}`); code != http.StatusForbidden {
		t.Fatalf("webhook palsu: %d, want 403", code)
	}

	// list transaksi ber-envelope + meta
	code, body = request(t, "GET", "/transactions?page=1&limit=10", staffToken, "")
	if code != http.StatusOK {
		t.Fatalf("list transaksi: %d", code)
	}
	if body["meta"].(map[string]any)["total"] != 1.0 {
		t.Fatalf("meta total = %v, want 1", body["meta"])
	}

	// delete transaksi: staff ditolak, admin boleh
	if code, _ = request(t, "DELETE", "/transactions/"+txID, staffToken, ""); code != http.StatusForbidden {
		t.Fatalf("delete oleh staff: %d, want 403", code)
	}
	if code, _ = request(t, "DELETE", "/transactions/"+txID, adminToken, ""); code != http.StatusOK {
		t.Fatalf("delete oleh admin: %d, want 200", code)
	}

	// laporan (admin)
	if code, body = request(t, "GET", "/reports/summary", adminToken, ""); code != http.StatusOK {
		t.Fatalf("reports: %d %v", code, body)
	}
}
