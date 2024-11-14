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

// Fungsi untuk menambah inventory baru
func AddInventory(w http.ResponseWriter, r *http.Request) {
	var inventory models.Inventory
	// Decode JSON body dari request ke dalam struct inventory
	if err := json.NewDecoder(r.Body).Decode(&inventory); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Simpan inventory ke database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Insert inventory ke dalam MongoDB
	_, err := config.InventoryCollection.InsertOne(ctx, inventory)
	if err != nil {
		http.Error(w, "Failed to create inventory", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Inventory added successfully"})
}

// Fungsi untuk mendapatkan daftar semua inventory
func GetAllInventory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query untuk mendapatkan semua inventory
	cursor, err := config.InventoryCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch inventory", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var inventory []models.Inventory
	for cursor.Next(ctx) {
		var item models.Inventory
		if err := cursor.Decode(&item); err != nil {
			http.Error(w, "Failed to decode inventory data", http.StatusInternalServerError)
			return
		}
		inventory = append(inventory, item)
	}

	if err := cursor.Err(); err != nil {
		http.Error(w, "Cursor error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventory)
}

// Fungsi untuk mendapatkan inventory berdasarkan ID
func GetInventoryByID(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari URL parameter
	inventoryID := r.URL.Query().Get("id")
	if inventoryID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Convert ID menjadi ObjectID MongoDB
	id, err := primitive.ObjectIDFromHex(inventoryID)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Query MongoDB untuk mencari inventory berdasarkan ID
	var item models.Inventory
	err = config.InventoryCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&item)
	if err != nil {
		http.Error(w, "Inventory not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// Fungsi untuk memperbarui data inventory
func UpdateInventory(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari URL parameter
	inventoryID := r.URL.Query().Get("id")
	if inventoryID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Convert ID menjadi ObjectID MongoDB
	id, err := primitive.ObjectIDFromHex(inventoryID)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Decode data inventory yang baru dari request body
	var updatedInventory models.Inventory
	if err := json.NewDecoder(r.Body).Decode(&updatedInventory); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Update data inventory di MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"namaProduk":   updatedInventory.NamaProduk,
			"deskripsi":    updatedInventory.Deskripsi,
			"jumlahStok":   updatedInventory.JumlahStok,
			"harga":        updatedInventory.Harga,
			"tanggalMasuk": updatedInventory.TanggalMasuk,
		},
	}

	_, err = config.InventoryCollection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		http.Error(w, "Failed to update inventory", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Inventory updated successfully"})
}

// Fungsi untuk menghapus inventory berdasarkan ID
func DeleteInventory(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari URL parameter
	inventoryID := r.URL.Query().Get("id")
	if inventoryID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Convert ID menjadi ObjectID MongoDB
	id, err := primitive.ObjectIDFromHex(inventoryID)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Hapus inventory dari MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = config.InventoryCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, "Failed to delete inventory", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Inventory deleted successfully"})
}
