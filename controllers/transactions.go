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
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}
	customerID, err := primitive.ObjectIDFromHex(req.CustomerID)
	if err != nil {
		http.Error(w, "customerId tidak valid", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "Pesanan minimal satu item", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var customer models.Customer
	if err := config.CustomerCollection.FindOne(ctx, bson.M{"_id": customerID}).Decode(&customer); err != nil {
		http.Error(w, "Customer tidak ditemukan", http.StatusBadRequest)
		return
	}

	items, total, err := recomputeItems(ctx, req.Items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, "Gagal menyimpan transaksi", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

func GetTransactions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := config.TransactionCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Gagal mendapatkan data transaksi", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		http.Error(w, "Gagal membaca data transaksi", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	objID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var transaction models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&transaction); err != nil {
		http.Error(w, "Transaksi tidak ditemukan", http.StatusNotFound)
		return
	}

	var customer models.Customer
	if config.CustomerCollection.FindOne(ctx, bson.M{"_id": transaction.CustomerID}).Decode(&customer) == nil {
		transaction.Customer = customer
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

// UpdateTransaction hanya untuk koreksi item saat pesanan masih pending;
// status, tanggal, dan data pembayaran tidak bisa ditimpa lewat endpoint ini.
func UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	objID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	var req struct {
		CustomerID string                   `json:"customerId"`
		Items      []models.TransactionItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "Pesanan minimal satu item", http.StatusBadRequest)
		return
	}
	customerID, err := primitive.ObjectIDFromHex(req.CustomerID)
	if err != nil {
		http.Error(w, "customerId tidak valid", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existing models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existing); err != nil {
		http.Error(w, "Transaksi tidak ditemukan", http.StatusNotFound)
		return
	}
	if existing.Status != "pending" {
		http.Error(w, "Pesanan sudah diproses dan tidak bisa diubah", http.StatusBadRequest)
		return
	}

	var customer models.Customer
	if err := config.CustomerCollection.FindOne(ctx, bson.M{"_id": customerID}).Decode(&customer); err != nil {
		http.Error(w, "Customer tidak ditemukan", http.StatusBadRequest)
		return
	}

	items, total, err := recomputeItems(ctx, req.Items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	update := bson.M{"$set": bson.M{
		"customerId":  customerID,
		"customer":    customer,
		"items":       items,
		"totalAmount": total,
	}}
	if _, err := config.TransactionCollection.UpdateOne(ctx, bson.M{"_id": objID}, update); err != nil {
		http.Error(w, "Gagal memperbarui transaksi", http.StatusInternalServerError)
		return
	}

	var updated models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updated); err != nil {
		http.Error(w, "Gagal mendapatkan transaksi setelah update", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func UpdateTransactionStatus(w http.ResponseWriter, r *http.Request) {
	objID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var transaction models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&transaction); err != nil {
		http.Error(w, "Transaksi tidak ditemukan", http.StatusNotFound)
		return
	}
	if !services.CanTransition(transaction.Status, req.Status) {
		http.Error(w, fmt.Sprintf("Transisi status %s ke %s tidak valid", transaction.Status, req.Status), http.StatusBadRequest)
		return
	}

	claims, _ := middleware.ClaimsFromContext(r.Context())
	update := bson.M{
		"$set":   bson.M{"status": req.Status},
		"$push":  bson.M{"statusHistory": models.StatusEvent{Status: req.Status, At: time.Now(), By: claims.Username}},
	}
	if _, err := config.TransactionCollection.UpdateOne(ctx, bson.M{"_id": objID}, update); err != nil {
		http.Error(w, "Gagal memperbarui status", http.StatusInternalServerError)
		return
	}

	var updated models.Transaction
	if err := config.TransactionCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updated); err != nil {
		http.Error(w, "Gagal mendapatkan transaksi setelah update", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	objID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err = config.TransactionCollection.DeleteOne(ctx, bson.M{"_id": objID}); err != nil {
		http.Error(w, "Gagal menghapus transaksi", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Transaksi berhasil dihapus"})
}
