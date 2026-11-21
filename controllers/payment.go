package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"laundry-pos/config"
	"laundry-pos/models"
	"laundry-pos/services"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/veritrans/go-midtrans"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// withTransaction menjalankan operasi multi-dokumen secara atomik.
// Butuh replica set; MongoDB Atlas memang replica set.
func withTransaction(ctx context.Context, fn func(mongo.SessionContext) error) error {
	sess, err := config.Client.StartSession()
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)
	_, err = sess.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		return nil, fn(sc)
	})
	return err
}

// CreatePayment membuat pembayaran Midtrans Snap untuk satu transaksi.
func CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TransactionID primitive.ObjectID `json:"transactionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TransactionID.IsZero() {
		http.Error(w, "transactionId tidak valid", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var transaction models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": req.TransactionID}).Decode(&transaction); err != nil {
		http.Error(w, "Transaksi tidak ditemukan", http.StatusNotFound)
		return
	}
	if transaction.PaymentStatus == "paid" {
		http.Error(w, "Pesanan sudah dibayar", http.StatusConflict)
		return
	}

	// Pakai ulang Snap yang masih pending supaya tidak dobel tagihan.
	var existing models.Payment
	err := config.PaymentCollection.FindOne(ctx, bson.M{
		"transactionId": req.TransactionID,
		"status":        "pending",
		"payment_method": "midtrans",
	}).Decode(&existing)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"snap_url": existing.SnapURL,
			"order_id": existing.OrderID,
		})
		return
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		http.Error(w, "Gagal mengambil data pembayaran", http.StatusInternalServerError)
		return
	}

	payment := models.Payment{
		ID:            primitive.NewObjectID(),
		TransactionID: transaction.ID,
		OrderID:       uuid.NewString(),
		GrossAmount:   transaction.TotalAmount,
		Status:        "pending",
		PaymentMethod: "midtrans",
		CreatedAt:     time.Now(),
	}

	snapGateway := midtrans.SnapGateway{Client: *services.MidtransClient()}
	snapResp, err := snapGateway.GetToken(&midtrans.SnapReq{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  payment.OrderID,
			GrossAmt: int64(payment.GrossAmount),
		},
	})
	if err != nil {
		http.Error(w, "Gagal membuat pembayaran Midtrans", http.StatusBadGateway)
		return
	}
	payment.SnapURL = snapResp.RedirectURL

	if _, err := config.PaymentCollection.InsertOne(ctx, payment); err != nil {
		http.Error(w, "Gagal menyimpan pembayaran", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"snap_url": payment.SnapURL,
		"order_id": payment.OrderID,
	})
}

// CreateCashPayment menandai pesanan lunas via tunai dan menghitung kembalian.
func CreateCashPayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TransactionID primitive.ObjectID `json:"transactionId"`
		AmountPaid    float64            `json:"amountPaid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TransactionID.IsZero() {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var transaction models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": req.TransactionID}).Decode(&transaction); err != nil {
		http.Error(w, "Transaksi tidak ditemukan", http.StatusNotFound)
		return
	}
	if transaction.PaymentStatus == "paid" {
		http.Error(w, "Pesanan sudah dibayar", http.StatusConflict)
		return
	}
	if req.AmountPaid < transaction.TotalAmount {
		http.Error(w, "Uang diterima kurang dari total", http.StatusBadRequest)
		return
	}

	now := time.Now()
	payment := models.Payment{
		ID:            primitive.NewObjectID(),
		TransactionID: transaction.ID,
		OrderID:       uuid.NewString(),
		GrossAmount:   transaction.TotalAmount,
		AmountPaid:    req.AmountPaid,
		Change:        req.AmountPaid - transaction.TotalAmount,
		Status:        "paid",
		PaymentMethod: "cash",
		CreatedAt:     now,
		PaidAt:        now,
	}

	// Filter paymentStatus mencegah dua kasir bayar bersamaan; yang kalah
	// menabrak ModifiedCount 0 dan transaksi Mongo di-roll back.
	err := withTransaction(ctx, func(sc mongo.SessionContext) error {
		res, err := config.TransactionCollection.UpdateOne(sc,
			bson.M{"_id": transaction.ID, "paymentStatus": bson.M{"$ne": "paid"}},
			bson.M{"$set": bson.M{"paymentMethod": "cash", "paymentStatus": "paid", "paidAt": now}},
		)
		if err != nil {
			return err
		}
		if res.ModifiedCount == 0 {
			return errors.New("pesanan sudah dibayar")
		}
		_, err = config.PaymentCollection.InsertOne(sc, payment)
		return err
	})
	if err != nil {
		http.Error(w, "Pesanan sudah dibayar", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

// WebhookHandler menerima notifikasi Midtrans. Wajib public, keamanannya lewat
// verifikasi signature SHA512.
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	orderID, _ := payload["order_id"].(string)
	statusCode, _ := payload["status_code"].(string)
	grossAmount, _ := payload["gross_amount"].(string)
	signature, _ := payload["signature_key"].(string)
	rawStatus, _ := payload["transaction_status"].(string)
	paymentType, _ := payload["payment_type"].(string)

	if !services.VerifyMidtransSignature(orderID, statusCode, grossAmount, os.Getenv("MIDTRANS_SERVER_KEY"), signature) {
		http.Error(w, "Signature tidak valid", http.StatusForbidden)
		return
	}

	status := services.MapMidtransStatus(rawStatus)
	if status == "" {
		// Status Midtrans yang tidak kita kenal: terima, jangan error,
		// supaya Midtrans tidak terus mengulang notifikasi yang sama.
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var payment models.Payment
	if err := config.PaymentCollection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&payment); err != nil {
		http.Error(w, "Pembayaran tidak ditemukan", http.StatusNotFound)
		return
	}
	if payment.Status == status {
		w.WriteHeader(http.StatusOK) // notifikasi duplikat, sudah diproses
		return
	}

	err := withTransaction(ctx, func(sc mongo.SessionContext) error {
		set := bson.M{"status": status, "midtrans_raw": rawStatus}
		if paymentType != "" {
			set["payment_method"] = paymentType
		}
		if status == "paid" {
			set["paid_at"] = time.Now()
		}
		if _, err := config.PaymentCollection.UpdateOne(sc, bson.M{"_id": payment.ID}, bson.M{"$set": set}); err != nil {
			return err
		}
		if status == "paid" {
			_, err := config.TransactionCollection.UpdateOne(sc,
				bson.M{"_id": payment.TransactionID, "paymentStatus": bson.M{"$ne": "paid"}},
				bson.M{"$set": bson.M{"paymentMethod": "midtrans", "paymentStatus": "paid", "paidAt": time.Now()}},
			)
			return err
		}
		return nil
	})
	if err != nil {
		// 500 supaya Midtrans mengulang notifikasi; idempoten di atas aman.
		http.Error(w, "Gagal memperbarui pembayaran", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetPaymentByOrderID mengembalikan satu pembayaran berdasarkan order_id.
func GetPaymentByOrderID(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("orderID")
	if orderID == "" {
		http.Error(w, "Order ID tidak disediakan", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var payment models.Payment
	if err := config.PaymentCollection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&payment); err != nil {
		http.Error(w, "Pembayaran tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

func GetAllPayments(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := config.PaymentCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Gagal mengambil data pembayaran", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var payments []models.Payment
	if err := cursor.All(ctx, &payments); err != nil {
		http.Error(w, "Gagal memproses data pembayaran", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}
