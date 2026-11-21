package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"laundry-pos/config"
	"laundry-pos/models"
	"laundry-pos/services"
	"laundry-pos/utils"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/veritrans/go-midtrans"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// errAlreadyPaid menandai race dua kasir bayar bersamaan (bukan error sistem).
var errAlreadyPaid = errors.New("pesanan sudah dibayar")

// ensureSnapToken mengembalikan token Snap; payment lama sebelum fitur popup
// tidak menyimpan token, jadi minta ulang ke Midtrans dengan order_id sama
// (Snap mengembalikan token yang sama untuk order_id yang masih aktif).
func ensureSnapToken(p *models.Payment) error {
	if p.SnapToken != "" {
		return nil
	}
	snapGateway := midtrans.SnapGateway{Client: *services.MidtransClient()}
	snapResp, err := snapGateway.GetToken(&midtrans.SnapReq{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  p.OrderID,
			GrossAmt: int64(p.GrossAmount),
		},
	})
	if err != nil {
		return err
	}
	p.SnapToken = snapResp.Token
	p.SnapURL = snapResp.RedirectURL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, _ = config.PaymentCollection.UpdateOne(ctx, bson.M{"_id": p.ID},
		bson.M{"$set": bson.M{"snap_token": p.SnapToken, "snap_url": p.SnapURL}})
	return nil
}

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
		utils.Err(w, http.StatusBadRequest, "transactionId tidak valid")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var transaction models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": req.TransactionID}).Decode(&transaction); err != nil {
		utils.Err(w, http.StatusNotFound, "Transaksi tidak ditemukan")
		return
	}
	if transaction.PaymentStatus == "paid" {
		utils.Err(w, http.StatusConflict, "Pesanan sudah dibayar")
		return
	}
	if transaction.Status == "batal" {
		utils.Err(w, http.StatusBadRequest, "Pesanan sudah dibatalkan")
		return
	}

	// Pakai ulang Snap yang masih pending supaya tidak dobel tagihan.
	var existing models.Payment
	err := config.PaymentCollection.FindOne(ctx, bson.M{
		"transactionId":  req.TransactionID,
		"status":         "pending",
		"payment_method": "midtrans",
	}).Decode(&existing)
	if err == nil {
		if err := ensureSnapToken(&existing); err != nil {
			utils.Err(w, http.StatusBadGateway, "Gagal membuat pembayaran Midtrans")
			return
		}
		utils.JSON(w, http.StatusOK, map[string]string{
			"snap_url":   existing.SnapURL,
			"snap_token": existing.SnapToken,
			"order_id":   existing.OrderID,
		})
		return
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		utils.Err(w, http.StatusInternalServerError, "Gagal mengambil data pembayaran")
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
		utils.Err(w, http.StatusBadGateway, "Gagal membuat pembayaran Midtrans")
		return
	}
	payment.SnapURL = snapResp.RedirectURL
	payment.SnapToken = snapResp.Token

	// Partial unique index menahan dua Snap pending untuk transaksi yang sama;
	// yang kalah balik memakai tagihan milik pemenang.
	if _, err := config.PaymentCollection.InsertOne(ctx, payment); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			var winner models.Payment
			if err := config.PaymentCollection.FindOne(ctx, bson.M{
				"transactionId":  req.TransactionID,
				"status":         "pending",
				"payment_method": "midtrans",
			}).Decode(&winner); err == nil {
				if err := ensureSnapToken(&winner); err != nil {
					utils.Err(w, http.StatusBadGateway, "Gagal membuat pembayaran Midtrans")
					return
				}
				utils.JSON(w, http.StatusOK, map[string]string{
					"snap_url":   winner.SnapURL,
					"snap_token": winner.SnapToken,
					"order_id":   winner.OrderID,
				})
				return
			}
		}
		utils.Err(w, http.StatusInternalServerError, "Gagal menyimpan pembayaran")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]string{
		"snap_url":   payment.SnapURL,
		"snap_token": payment.SnapToken,
		"order_id":   payment.OrderID,
	})
}

// CreateCashPayment menandai pesanan lunas via tunai dan menghitung kembalian.
func CreateCashPayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TransactionID primitive.ObjectID `json:"transactionId"`
		AmountPaid    float64            `json:"amountPaid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TransactionID.IsZero() {
		utils.Err(w, http.StatusBadRequest, "Input tidak valid")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var transaction models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": req.TransactionID}).Decode(&transaction); err != nil {
		utils.Err(w, http.StatusNotFound, "Transaksi tidak ditemukan")
		return
	}
	if transaction.PaymentStatus == "paid" {
		utils.Err(w, http.StatusConflict, "Pesanan sudah dibayar")
		return
	}
	if transaction.Status == "batal" {
		utils.Err(w, http.StatusBadRequest, "Pesanan sudah dibatalkan")
		return
	}
	if req.AmountPaid < transaction.TotalAmount {
		utils.Err(w, http.StatusBadRequest, "Uang diterima kurang dari total")
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
			return errAlreadyPaid
		}
		_, err = config.PaymentCollection.InsertOne(sc, payment)
		return err
	})
	if err != nil {
		if errors.Is(err, errAlreadyPaid) {
			utils.Err(w, http.StatusConflict, "Pesanan sudah dibayar")
		} else {
			utils.Err(w, http.StatusInternalServerError, "Gagal menyimpan pembayaran")
		}
		return
	}

	utils.JSON(w, http.StatusCreated, payment)
}

// WebhookHandler menerima notifikasi Midtrans. Wajib public, keamanannya lewat
// verifikasi signature SHA512.
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Err(w, http.StatusBadRequest, "Invalid payload")
		return
	}

	orderID, _ := payload["order_id"].(string)
	statusCode, _ := payload["status_code"].(string)
	grossAmount, _ := payload["gross_amount"].(string)
	signature, _ := payload["signature_key"].(string)
	rawStatus, _ := payload["transaction_status"].(string)
	paymentType, _ := payload["payment_type"].(string)

	if !services.VerifyMidtransSignature(orderID, statusCode, grossAmount, os.Getenv("MIDTRANS_SERVER_KEY"), signature) {
		utils.Err(w, http.StatusForbidden, "Signature tidak valid")
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
		utils.Err(w, http.StatusNotFound, "Pembayaran tidak ditemukan")
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
		utils.Err(w, http.StatusInternalServerError, "Gagal memperbarui pembayaran")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetPaymentByOrderID mengembalikan satu pembayaran berdasarkan order_id.
func GetPaymentByOrderID(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("orderID")
	if orderID == "" {
		utils.Err(w, http.StatusBadRequest, "Order ID tidak disediakan")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var payment models.Payment
	if err := config.PaymentCollection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&payment); err != nil {
		utils.Err(w, http.StatusNotFound, "Pembayaran tidak ditemukan")
		return
	}

	utils.JSON(w, http.StatusOK, payment)
}

func GetAllPayments(w http.ResponseWriter, r *http.Request) {
	page, limit := utils.PageParams(r)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	total, err := config.PaymentCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal menghitung pembayaran")
		return
	}

	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})
	cursor, err := config.PaymentCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal mengambil pembayaran")
		return
	}
	defer cursor.Close(ctx)

	var payments []models.Payment
	if err := cursor.All(ctx, &payments); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal memproses pembayaran")
		return
	}

	utils.Page(w, http.StatusOK, payments, map[string]any{"page": page, "limit": limit, "total": total})
}
