package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	JSON(rec, http.StatusCreated, map[string]string{"a": "b"})

	if rec.Code != http.StatusCreated {
		t.Fatalf("code = %d, want 201", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q", ct)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["success"] != true {
		t.Fatal("success harus true")
	}
	data := body["data"].(map[string]any)
	if data["a"] != "b" {
		t.Fatalf("data = %v", data)
	}
}

func TestErrEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	Err(rec, http.StatusBadRequest, "Input tidak valid")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["success"] != false || body["message"] != "Input tidak valid" {
		t.Fatalf("body = %v", body)
	}
}

func TestPageEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	meta := map[string]any{"page": 1, "limit": 20, "total": 42}
	Page(rec, http.StatusOK, []string{"x"}, meta)

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["success"] != true {
		t.Fatal("success harus true")
	}
	m := body["meta"].(map[string]any)
	if m["total"] != 42.0 {
		t.Fatalf("meta = %v", m)
	}
}

func TestPageParams(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/x?page=3&limit=50", nil)
	page, limit := PageParams(r)
	if page != 3 || limit != 50 {
		t.Fatalf("page=%d limit=%d, want 3/50", page, limit)
	}

	r = httptest.NewRequest(http.MethodGet, "/x", nil)
	page, limit = PageParams(r)
	if page != 1 || limit != 20 {
		t.Fatalf("default page=%d limit=%d, want 1/20", page, limit)
	}

	r = httptest.NewRequest(http.MethodGet, "/x?page=0&limit=5000", nil)
	page, limit = PageParams(r)
	if page < 1 || limit > 100 {
		t.Fatalf("page=%d limit=%d melewati batas", page, limit)
	}
}
