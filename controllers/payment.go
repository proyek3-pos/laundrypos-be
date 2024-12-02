package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"laundry-pos/config"
	"laundry-pos/models"
	"laundry-pos/services"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/veritrans/go-midtrans"
)

// Struktur permintaan pembayaran
type PaymentRequest struct {
	OrderID     string  `json:"order_id"`
	GrossAmount float64 `json:"gross_amount"`
	Customer    struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	} `json:"customer"`
}

// Fungsi untuk membuat pembayaran menggunakan Midtrans dan menyimpannya di MongoDB
func CreatePayment(w http.ResponseWriter, r *http.Request) {
	var paymentReq PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&paymentReq); err != nil {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}

	// Generate UUID untuk Order ID jika tidak disertakan
	if paymentReq.OrderID == "" {
		paymentReq.OrderID = uuid.New().String() // Membuat order ID unik
	}

	// Inisialisasi Midtrans Client
	midtransClient := services.MidtransClient()

	// Inisialisasi Snap Gateway
	snapGateway := midtrans.SnapGateway{
		Client: *midtransClient, // Dereference pointer menjadi nilai
	}

	// Buat permintaan transaksi Snap
	snapReq := &midtrans.SnapReq{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  paymentReq.OrderID,
			GrossAmt: int64(paymentReq.GrossAmount), // Konversi ke int64 dan perbaiki penghitungan grossAmount
		},
		CustomerDetail: &midtrans.CustDetail{
			FName: paymentReq.Customer.Name,
			Email: paymentReq.Customer.Email,
			Phone: paymentReq.Customer.Phone,
		},
	}

	// Dapatkan Snap URL
	snapResp, err := snapGateway.GetToken(snapReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal membuat transaksi: %v", err), http.StatusInternalServerError)
		return
	}

	// Simpan informasi transaksi pembayaran ke MongoDB
	payment := models.Payment{
		OrderID:     paymentReq.OrderID,
		GrossAmount: paymentReq.GrossAmount,
		SnapURL:     snapResp.RedirectURL,
		Customer: models.PaymentCustomer{
			Name:  paymentReq.Customer.Name,
			Email: paymentReq.Customer.Email,
			Phone: paymentReq.Customer.Phone,
		},
		Status: "Pending", // Status pembayaran bisa diset ke 'Pending' dulu
	}

	// Menyimpan transaksi pembayaran ke MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = config.PaymentCollection.InsertOne(ctx, payment)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal menyimpan data pembayaran: %v", err), http.StatusInternalServerError)
		return
	}

	// Kembalikan Snap URL ke client
	json.NewEncoder(w).Encode(map[string]string{
		"snap_url": snapResp.RedirectURL,
		"order_id": paymentReq.OrderID,
	})
}
