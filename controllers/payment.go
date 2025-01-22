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
	"go.mongodb.org/mongo-driver/bson"
)

// Membuat pembayaran menggunakan Midtrans
func CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var paymentReq models.Payment
	if err := json.NewDecoder(r.Body).Decode(&paymentReq); err != nil {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}

	// Pastikan transaction_id tersedia
	if paymentReq.TransactionID.IsZero() {
		http.Error(w, "Transaction ID tidak disediakan", http.StatusBadRequest)
		return
	}

	// Ambil transaksi dari database berdasarkan TransactionID yang diberikan
	var transaction models.Transaction
	err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": paymentReq.TransactionID}).Decode(&transaction)
	if err != nil {
		http.Error(w, "Transaksi tidak ditemukan", http.StatusNotFound)
		return
	}

	// Menghitung gross_amount dalam cent
	// Jangan kalikan dua kali! Pastikan ini adalah dalam rupiah saja (misalnya 16000)
	// Jika transaksi sudah dalam rupiah, kirim langsung dengan mengalikannya hanya di bagian snapReq.

	paymentReq.GrossAmount = transaction.TotalAmount // Jangan kalikan dengan 100 di sini, ini sudah dalam rupiah

	// Generate OrderID jika belum disediakan
	if paymentReq.OrderID == "" {
		paymentReq.OrderID = uuid.New().String()
	}

	// Inisialisasi Midtrans Client dan Snap Gateway
	midtransClient := services.MidtransClient()
	snapGateway := midtrans.SnapGateway{Client: *midtransClient}

	// Membuat request ke Midtrans dengan data transaksi
	snapReq := &midtrans.SnapReq{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  paymentReq.OrderID,
			GrossAmt: int64(paymentReq.GrossAmount), 
		},
	}

	snapResp, err := snapGateway.GetToken(snapReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal membuat pembayaran: %v", err), http.StatusInternalServerError)
		return
	}

	// Tambahkan snap_url dan status pembayaran
	paymentReq.SnapURL = snapResp.RedirectURL
	paymentReq.Status = "Pending"
	paymentReq.CreatedAt = time.Now()

	// Simpan pembayaran
	_, err = config.PaymentCollection.InsertOne(ctx, paymentReq)
	if err != nil {
		http.Error(w, "Gagal menyimpan pembayaran", http.StatusInternalServerError)
		return
	}

	// Ambil informasi customer dari database berdasarkan CustomerID yang ada di transaksi
	var customer models.Customer
	err = config.CustomerCollection.FindOne(ctx, bson.M{"_id": transaction.CustomerID}).Decode(&customer)
	if err != nil {
		http.Error(w, "Customer tidak ditemukan", http.StatusNotFound)
		return
	}

	// Kirim data konfirmasi pembayaran yang diperlukan
	confirmationData := map[string]interface{}{
		"fullName":     customer.FullName,
		"phoneNumber":  customer.PhoneNumber,
		"email":        customer.Email,
		"service_name": transaction.Items[0].Service.ServiceName,
		"quantity":     transaction.Items[0].Quantity,
		"total_amount": transaction.TotalAmount,
	}

	// Kirim response dengan snap_url, order_id, dan konfirmasi data
	response := map[string]interface{}{
		"snap_url":          paymentReq.SnapURL,
		"order_id":          paymentReq.OrderID,
		"confirmation_data": confirmationData, // Data konfirmasi transaksi
	}

	json.NewEncoder(w).Encode(response)
}

// WebhookHandler menangani notifikasi dari Midtrans
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
    var notificationPayload map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&notificationPayload); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }

    // Ambil Order ID dari payload
    orderID, ok := notificationPayload["order_id"].(string)
    if !ok {
        http.Error(w, "Order ID tidak valid", http.StatusBadRequest)
        return
    }

    // Ambil status pembayaran dari payload
    transactionStatus, ok := notificationPayload["transaction_status"].(string)
    if !ok {
        http.Error(w, "Status transaksi tidak valid", http.StatusBadRequest)
        return
    }

    ctx := context.Background()

    // Update status pembayaran di database
    update := bson.M{"$set": bson.M{"status": transactionStatus}}
    _, err := config.PaymentCollection.UpdateOne(ctx, bson.M{"order_id": orderID}, update)
    if err != nil {
        http.Error(w, "Gagal memperbarui status pembayaran", http.StatusInternalServerError)
        return
    }

    // Response ke Midtrans bahwa notifikasi diterima
    w.WriteHeader(http.StatusOK)
}

// GetPaymentByOrderID mendapatkan data pembayaran berdasarkan OrderID
func GetPaymentByOrderID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Ambil OrderID dari query parameter
	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		http.Error(w, "Order ID tidak disediakan", http.StatusBadRequest)
		return
	}

	// Cari pembayaran di database berdasarkan OrderID
	var payment models.Payment
	err := config.PaymentCollection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&payment)
	if err != nil {
		http.Error(w, "Pembayaran tidak ditemukan", http.StatusNotFound)
		return
	}

	// Kembalikan data pembayaran dalam bentuk JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

// GetAllPayments mendapatkan semua data pembayaran
func GetAllPayments(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Cari semua pembayaran di database
	cursor, err := config.PaymentCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Gagal mengambil data pembayaran", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Parse data pembayaran ke dalam slice
	var payments []models.Payment
	for cursor.Next(ctx) {
		var payment models.Payment
		if err := cursor.Decode(&payment); err != nil {
			http.Error(w, "Gagal memproses data pembayaran", http.StatusInternalServerError)
			return
		}
		payments = append(payments, payment)
	}

	// Kembalikan semua data pembayaran dalam bentuk JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}
