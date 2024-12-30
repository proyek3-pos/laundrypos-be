package controllers

import (
	"context"
	"encoding/json"
	"laundry-pos/config"
	"laundry-pos/models"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Fungsi untuk membuat transaksi baru
func CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var transaction models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}

	// Validasi keberadaan CustomerID
	var customer models.Customer
	err := config.CustomerCollection.FindOne(context.TODO(), bson.M{"_id": transaction.CustomerID}).Decode(&customer)
	if err != nil {
		http.Error(w, "Customer tidak ditemukan", http.StatusBadRequest)
		return
	}

	transaction.ID = primitive.NewObjectID()
	transaction.TransactionDate = time.Now()
	transaction.Status = "Pending"
	transaction.Customer = customer


	var totalAmount float64

	for i, item := range transaction.Items {
		var service models.Service
		err := config.ServiceCollection.FindOne(context.TODO(), bson.M{"_id": item.ServiceID}).Decode(&service)
		if err != nil {
			http.Error(w, "Layanan tidak ditemukan", http.StatusBadRequest)
			return
		}

		transaction.Items[i].ID = primitive.NewObjectID()
		transaction.Items[i].UnitPrice = service.UnitPrice
		transaction.Items[i].Service = service
		transaction.Items[i].TotalPrice = float64(item.Quantity) * service.UnitPrice
		totalAmount += transaction.Items[i].TotalPrice
	}

	transaction.TotalAmount = totalAmount

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = config.TransactionCollection.InsertOne(ctx, transaction)
	if err != nil {
		http.Error(w, "Gagal menyimpan transaksi", http.StatusInternalServerError)
		return
	}

	// Tambahkan data customer ke response
	transaction.Customer = customer

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

// Fungsi untuk mendapatkan semua transaksi
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

	// Populate data customer untuk setiap transaksi
	for i, transaction := range transactions {
		var customer models.Customer
		err := config.CustomerCollection.FindOne(ctx, bson.M{"_id": transaction.CustomerID}).Decode(&customer)
		if err == nil {
			transactions[i].Customer = customer
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// Fungsi untuk mendapatkan transaksi berdasarkan ID
func GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID tidak disediakan", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var transaction models.Transaction
	err = config.TransactionCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&transaction)
	if err != nil {
		http.Error(w, "Transaksi tidak ditemukan", http.StatusNotFound)
		return
	}

	var customer models.Customer
	err = config.CustomerCollection.FindOne(ctx, bson.M{"_id": transaction.CustomerID}).Decode(&customer)
	if err == nil {
		transaction.Customer = customer
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

// Fungsi untuk memperbarui transaksi
func UpdateTransaction(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    if id == "" {
        http.Error(w, "ID tidak disediakan", http.StatusBadRequest)
        return
    }

    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        http.Error(w, "ID tidak valid", http.StatusBadRequest)
        return
    }

    var transaction models.Transaction
    if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
        http.Error(w, "Input tidak valid", http.StatusBadRequest)
        return
    }

    var totalAmount float64
    for i, item := range transaction.Items {

        var service models.Service
        err := config.ServiceCollection.FindOne(context.TODO(), bson.M{"_id": item.ServiceID}).Decode(&service)
        if err != nil {
            http.Error(w, "Layanan tidak ditemukan", http.StatusBadRequest)
            return
        }
		
		transaction.Items[i].ID = item.ID  // Tetap gunakan ID yang ada
        transaction.Items[i].UnitPrice = service.UnitPrice
        transaction.Items[i].Service = service
        transaction.Items[i].TotalPrice = float64(item.Quantity) * service.UnitPrice
        totalAmount += transaction.Items[i].TotalPrice
    }

    transaction.TotalAmount = totalAmount
    transaction.TransactionDate = time.Now()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Update transaksi di database
    update := bson.M{
        "$set": transaction,
    }

    result, err := config.TransactionCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
    if err != nil {
        http.Error(w, "Gagal memperbarui transaksi: "+err.Error(), http.StatusInternalServerError)
        return
    }
    if result.ModifiedCount == 0 {
        http.Error(w, "Transaksi tidak ditemukan untuk diperbarui", http.StatusNotFound)
        return
    }

    // Ambil kembali data transaksi yang telah diperbarui
    err = config.TransactionCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&transaction)
    if err != nil {
        http.Error(w, "Gagal mendapatkan transaksi setelah update: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Tambahkan detail customer
    var customer models.Customer
    err = config.CustomerCollection.FindOne(ctx, bson.M{"_id": transaction.CustomerID}).Decode(&customer)
    if err == nil {
        transaction.Customer = customer
    }

    // Kirim response
    w.Header().Set("Content-Type", "application/json")
    response := map[string]interface{}{
        "message":     "Transaksi berhasil diperbarui",
        "transaction": transaction,
    }
    json.NewEncoder(w).Encode(response)
}



// Fungsi untuk menghapus transaksi
func DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID tidak disediakan", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = config.TransactionCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Gagal menghapus transaksi", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Transaksi berhasil dihapus",
		"deleted_id": objID.Hex(),
	})
}
