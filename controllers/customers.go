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

// Fungsi untuk menambah customer baru
func AddCustomer(w http.ResponseWriter, r *http.Request) {
	var customer models.Customer
	// Decode JSON body dari request ke dalam struct customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Simpan customer ke database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Insert customer ke dalam MongoDB
	_, err := config.CustomerCollection.InsertOne(ctx, customer)
	if err != nil {
		http.Error(w, "Failed to create customer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Customer added successfully"})
}

// Fungsi untuk mendapatkan daftar semua customers
func GetAllCustomers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query untuk mendapatkan semua customers
	cursor, err := config.CustomerCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch customers", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var customers []models.Customer
	for cursor.Next(ctx) {
		var customer models.Customer
		if err := cursor.Decode(&customer); err != nil {
			http.Error(w, "Failed to decode customer data", http.StatusInternalServerError)
			return
		}
		customers = append(customers, customer)
	}

	if err := cursor.Err(); err != nil {
		http.Error(w, "Cursor error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

// Fungsi untuk mendapatkan customer berdasarkan ID
func GetCustomerByID(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari URL parameter
	customerID := r.URL.Query().Get("id")
	if customerID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Convert ID menjadi ObjectID MongoDB
	id, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Query MongoDB untuk mencari customer berdasarkan ID
	var customer models.Customer
	err = config.CustomerCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&customer)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

// Fungsi untuk memperbarui data customer
func UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari URL parameter
	customerID := r.URL.Query().Get("id")
	if customerID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Convert ID menjadi ObjectID MongoDB
	id, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Decode data customer yang baru dari request body
	var updatedCustomer models.Customer
	if err := json.NewDecoder(r.Body).Decode(&updatedCustomer); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Update data customer di MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"fullName":   updatedCustomer.FullName,
			"phoneNumber": updatedCustomer.PhoneNumber,
			"email":      updatedCustomer.Email,
		},
	}

	_, err = config.CustomerCollection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		http.Error(w, "Failed to update customer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Customer updated successfully"})
}

// Fungsi untuk menghapus customer berdasarkan ID
func DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari URL parameter
	customerID := r.URL.Query().Get("id")
	if customerID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Convert ID menjadi ObjectID MongoDB
	id, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Hapus customer dari MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = config.CustomerCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, "Failed to delete customer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Customer deleted successfully"})
}
