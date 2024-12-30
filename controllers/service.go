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

// Fungsi untuk menambah layanan baru
func CreateService(w http.ResponseWriter, r *http.Request) {
	var service models.Service
	// Decode JSON body dari request ke dalam struct service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}

	// Generate ObjectID untuk layanan baru
	service.ID = primitive.NewObjectID()

	// Simpan service ke database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := config.ServiceCollection.InsertOne(ctx, service)
	if err != nil {
		http.Error(w, "Gagal menambahkan layanan", http.StatusInternalServerError)
		return
	}

	// Kirim response yang berisi pesan sukses dan data layanan yang baru ditambahkan
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "Layanan berhasil ditambahkan",
		"service": service,
	}
	json.NewEncoder(w).Encode(response)
}

// Dapatkan semua layanan
func GetAllServices(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query untuk mendapatkan semua layanan
	cursor, err := config.ServiceCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Gagal mendapatkan layanan", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var services []models.Service
	for cursor.Next(ctx) {
		var service models.Service
		if err := cursor.Decode(&service); err != nil {
			http.Error(w, "Gagal membaca layanan", http.StatusInternalServerError)
			return
		}
		services = append(services, service)
	}

	if err := cursor.Err(); err != nil {
		http.Error(w, "Cursor error", http.StatusInternalServerError)
		return
	}

	// Kirim daftar layanan
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// Dapatkan layanan berdasarkan ID
func GetServiceByID(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari URL parameter
	serviceID := r.URL.Query().Get("id")
	if serviceID == "" {
		http.Error(w, "ID tidak disediakan", http.StatusBadRequest)
		return
	}

	// Convert ID menjadi ObjectID MongoDB
	id, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	// Query MongoDB untuk mencari service berdasarkan ID
	var service models.Service
	err = config.ServiceCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&service)
	if err != nil {
		http.Error(w, "Layanan tidak ditemukan", http.StatusNotFound)
		return
	}

	// Kirim response service yang ditemukan
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

// Fungsi untuk memperbarui layanan
func UpdateService(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari URL parameter
	serviceID := r.URL.Query().Get("id")
	if serviceID == "" {
		http.Error(w, "ID tidak disediakan", http.StatusBadRequest)
		return
	}

	// Convert ID menjadi ObjectID MongoDB
	id, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	// Decode data service yang baru dari request body
	var updatedService models.Service
	if err := json.NewDecoder(r.Body).Decode(&updatedService); err != nil {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}

	// Set ID ke service sebelum update
	updatedService.ID = id

	// Update data service di MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"serviceName": updatedService.ServiceName,
			"description": updatedService.Description,
			"unitPrice":   updatedService.UnitPrice,
			"unit":        updatedService.Unit,
		},
	}

	_, err = config.ServiceCollection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		http.Error(w, "Gagal memperbarui layanan", http.StatusInternalServerError)
		return
	}

	// Kirim response dengan pesan sukses dan data yang diperbarui
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "Layanan berhasil diperbarui",
		"service": updatedService,
	}
	json.NewEncoder(w).Encode(response)
}


// Fungsi untuk menghapus layanan
func DeleteService(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari URL parameter
	serviceID := r.URL.Query().Get("id")
	if serviceID == "" {
		http.Error(w, "ID tidak disediakan", http.StatusBadRequest)
		return
	}

	// Convert ID menjadi ObjectID MongoDB
	id, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	// Hapus service dari MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = config.ServiceCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, "Gagal menghapus layanan", http.StatusInternalServerError)
		return
	}

	// Kirim response bahwa layanan berhasil dihapus dengan pesan dan ID yang dihapus
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message":    "Layanan berhasil dihapus",
		"deleted_id": id.Hex(),
	}
	json.NewEncoder(w).Encode(response)
}
