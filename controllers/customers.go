package controllers

import (
	"encoding/json"
	"laundry-pos/config"
	"laundry-pos/models"
	"net/http"

	"gorm.io/gorm"
)

// Fungsi untuk menambah customer baru
func AddCustomer(w http.ResponseWriter, r *http.Request) {
	var customer models.Customer
	// Decode JSON body dari request ke dalam struct customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validasi data customer (contoh: pastikan nama depan dan belakang ada)
	if customer.FirstName == "" || customer.LastName == "" {
		http.Error(w, "First Name and Last Name are required", http.StatusBadRequest)
		return
	}

	// Simpan customer ke database
	if err := config.DB.Create(&customer).Error; err != nil {
		http.Error(w, "Failed to create customer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Customer added successfully"})
}

// Fungsi untuk mendapatkan daftar semua customers
func GetAllCustomers(w http.ResponseWriter, r *http.Request) {
	var customers []models.Customer

	// Query untuk mendapatkan semua customers
	if err := config.DB.Find(&customers).Error; err != nil {
		http.Error(w, "Failed to fetch customers", http.StatusInternalServerError)
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

	// Convert ID menjadi UUID (menggunakan GORM UUID)
	var customer models.Customer
	err := config.DB.First(&customer, "id = ?", customerID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error fetching customer", http.StatusInternalServerError)
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

	// Decode data customer yang baru dari request body
	var updatedCustomer models.Customer
	if err := json.NewDecoder(r.Body).Decode(&updatedCustomer); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Update data customer di database
	if err := config.DB.Model(&models.Customer{}).Where("id = ?", customerID).Updates(updatedCustomer).Error; err != nil {
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

	// Hapus customer dari database
	if err := config.DB.Delete(&models.Customer{}, "id = ?", customerID).Error; err != nil {
		http.Error(w, "Failed to delete customer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Customer deleted successfully"})
}
