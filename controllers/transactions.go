package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"laundry-pos/config"
	"laundry-pos/middleware"
	"laundry-pos/models"
	"laundry-pos/services"
	"laundry-pos/utils"
	"net/http"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// recomputeItems memvalidasi item dan menghitung ulang harga dari data layanan;
// harga dan total dari client tidak pernah dipercaya.
func recomputeItems(ctx context.Context, in []models.TransactionItem) ([]models.TransactionItem, float64, error) {
	items := make([]models.TransactionItem, len(in))
	var total float64
	for i, item := range in {
		var service models.Service
		if err := config.ServiceCollection.FindOne(ctx, bson.M{"_id": item.ServiceID}).Decode(&service); err != nil {
			return nil, 0, fmt.Errorf("layanan tidak ditemukan")
		}
		if err := utils.ValidateQuantity(item.Quantity, service.Unit); err != nil {
			return nil, 0, err
		}
		item.ID = primitive.NewObjectID()
		item.UnitPrice = service.UnitPrice
		item.TotalPrice = item.Quantity * service.UnitPrice
		items[i] = item
		total += item.TotalPrice
	}
	return items, total, nil
}

func CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID string                   `json:"customerId"`
		Items      []models.TransactionItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Err(w, http.StatusBadRequest, "Input tidak valid")
		return
	}
	customerID, err := primitive.ObjectIDFromHex(req.CustomerID)
	if err != nil {
		utils.Err(w, http.StatusBadRequest, "customerId tidak valid")
		return
	}
	if len(req.Items) == 0 {
		utils.Err(w, http.StatusBadRequest, "Pesanan minimal satu item")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var customer models.Customer
	if err := config.CustomerCollection.FindOne(ctx, bson.M{"_id": customerID}).Decode(&customer); err != nil {
		utils.Err(w, http.StatusBadRequest, "Customer tidak ditemukan")
		return
	}

	items, total, err := recomputeItems(ctx, req.Items)
	if err != nil {
		utils.Err(w, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now()
	claims, _ := middleware.ClaimsFromContext(r.Context())
	transaction := models.Transaction{
		ID:              primitive.NewObjectID(),
		InvoiceNumber:   services.NewInvoiceNumber(now),
		CustomerID:      customerID,
		Customer:        customer,
		TransactionDate: now,
		Items:           items,
		TotalAmount:     total,
		PaymentStatus:   "unpaid",
		Status:          "pending",
		StatusHistory:   []models.StatusEvent{{Status: "pending", At: now, By: claims.Username}},
	}

	if _, err := config.TransactionCollection.InsertOne(ctx, transaction); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal menyimpan transaksi")
		return
	}

	utils.JSON(w, http.StatusCreated, transaction)
}

func GetTransactions(w http.ResponseWriter, r *http.Request) {
	page, limit := utils.PageParams(r)
	filter := bson.M{}
	if s := r.URL.Query().Get("status"); s != "" {
		filter["status"] = s
	}
	if q := r.URL.Query().Get("q"); q != "" {
		filter["invoiceNumber"] = bson.M{"$regex": regexp.QuoteMeta(q), "$options": "i"}
	}
	dateFilter := bson.M{}
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02", v, time.Local); err == nil {
			dateFilter["$gte"] = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02", v, time.Local); err == nil {
			dateFilter["$lt"] = t.AddDate(0, 0, 1)
		}
	}
	if len(dateFilter) > 0 {
		filter["transactionDate"] = dateFilter
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	total, err := config.TransactionCollection.CountDocuments(ctx, filter)
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal menghitung transaksi")
		return
	}

	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"transactionDate": -1})
	cursor, err := config.TransactionCollection.Find(ctx, filter, opts)
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal mendapatkan transaksi")
		return
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal membaca transaksi")
		return
	}

	// Satu query $in untuk semua customer, bukan FindOne per transaksi.
	ids := make([]primitive.ObjectID, 0, len(transactions))
	for _, t := range transactions {
		ids = append(ids, t.CustomerID)
	}
	customers := map[primitive.ObjectID]models.Customer{}
	if len(ids) > 0 {
		cCursor, err := config.CustomerCollection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
		if err == nil {
			for cCursor.Next(ctx) {
				var c models.Customer
				if cCursor.Decode(&c) == nil {
					customers[c.ID] = c
				}
			}
			cCursor.Close(ctx)
		}
	}
	for i, t := range transactions {
		if c, ok := customers[t.CustomerID]; ok {
			transactions[i].Customer = c
		}
	}

	utils.Page(w, http.StatusOK, transactions, map[string]any{"page": page, "limit": limit, "total": total})
}

func GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	objID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		utils.Err(w, http.StatusBadRequest, "ID tidak valid")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var transaction models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&transaction); err != nil {
		utils.Err(w, http.StatusNotFound, "Transaksi tidak ditemukan")
		return
	}

	var customer models.Customer
	if config.CustomerCollection.FindOne(ctx, bson.M{"_id": transaction.CustomerID}).Decode(&customer) == nil {
		transaction.Customer = customer
	}

	utils.JSON(w, http.StatusOK, transaction)
}

// UpdateTransaction hanya untuk koreksi item saat pesanan masih pending;
// status, tanggal, dan data pembayaran tidak bisa ditimpa lewat endpoint ini.
func UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	objID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		utils.Err(w, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var req struct {
		CustomerID string                   `json:"customerId"`
		Items      []models.TransactionItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Err(w, http.StatusBadRequest, "Input tidak valid")
		return
	}
	if len(req.Items) == 0 {
		utils.Err(w, http.StatusBadRequest, "Pesanan minimal satu item")
		return
	}
	customerID, err := primitive.ObjectIDFromHex(req.CustomerID)
	if err != nil {
		utils.Err(w, http.StatusBadRequest, "customerId tidak valid")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existing models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existing); err != nil {
		utils.Err(w, http.StatusNotFound, "Transaksi tidak ditemukan")
		return
	}
	if existing.Status != "pending" {
		utils.Err(w, http.StatusBadRequest, "Pesanan sudah diproses dan tidak bisa diubah")
		return
	}

	var customer models.Customer
	if err := config.CustomerCollection.FindOne(ctx, bson.M{"_id": customerID}).Decode(&customer); err != nil {
		utils.Err(w, http.StatusBadRequest, "Customer tidak ditemukan")
		return
	}

	items, total, err := recomputeItems(ctx, req.Items)
	if err != nil {
		utils.Err(w, http.StatusBadRequest, err.Error())
		return
	}

	update := bson.M{"$set": bson.M{
		"customerId":  customerID,
		"customer":    customer,
		"items":       items,
		"totalAmount": total,
	}}
	if _, err := config.TransactionCollection.UpdateOne(ctx, bson.M{"_id": objID}, update); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal memperbarui transaksi")
		return
	}

	var updated models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updated); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal mendapatkan transaksi setelah update")
		return
	}

	utils.JSON(w, http.StatusOK, updated)
}

func UpdateTransactionStatus(w http.ResponseWriter, r *http.Request) {
	objID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		utils.Err(w, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Err(w, http.StatusBadRequest, "Input tidak valid")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var transaction models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&transaction); err != nil {
		utils.Err(w, http.StatusNotFound, "Transaksi tidak ditemukan")
		return
	}
	if !services.CanTransition(transaction.Status, req.Status) {
		utils.Err(w, http.StatusBadRequest, fmt.Sprintf("Transisi status %s ke %s tidak valid", transaction.Status, req.Status))
		return
	}

	claims, _ := middleware.ClaimsFromContext(r.Context())
	update := bson.M{
		"$set":  bson.M{"status": req.Status},
		"$push": bson.M{"statusHistory": models.StatusEvent{Status: req.Status, At: time.Now(), By: claims.Username}},
	}
	if _, err := config.TransactionCollection.UpdateOne(ctx, bson.M{"_id": objID}, update); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal memperbarui status")
		return
	}

	var updated models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updated); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal mendapatkan transaksi setelah update")
		return
	}

	utils.JSON(w, http.StatusOK, updated)
}

func DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	objID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		utils.Err(w, http.StatusBadRequest, "ID tidak valid")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err = config.TransactionCollection.DeleteOne(ctx, bson.M{"_id": objID}); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal menghapus transaksi")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Transaksi dihapus"})
}
