package routes

import (
	"laundry-pos/controllers"
	"laundry-pos/middleware"
	"net/http"
)

func InitRoutes() *http.ServeMux {
	router := http.NewServeMux()

	// Rute Auth
	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.Register(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.Login(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Rute untuk customer dengan middleware untuk otentikasi
	router.Handle("/customers", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.AddCustomer(w, r) // Membuat customer baru
		case http.MethodGet:
			controllers.GetAllCustomers(w, r) // Mengambil semua data customer
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Rute untuk mengambil customer berdasarkan ID dengan middleware untuk otentikasi
	router.Handle("/customer-id", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			controllers.GetCustomerByID(w, r) // Mengambil data customer berdasarkan ID
		case http.MethodPut:
			controllers.UpdateCustomer(w, r) // Mengupdate data customer berdasarkan ID
		case http.MethodDelete:
			controllers.DeleteCustomer(w, r) // Menghapus data customer berdasarkan ID
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Rute untuk manajemen inventaris dengan middleware untuk otentikasi
	router.Handle("/inventory", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.AddInventory(w, r) // Menambah item ke dalam inventaris
		case http.MethodGet:
			controllers.GetAllInventory(w, r) // Mengambil semua data inventaris
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Rute untuk mengambil inventaris berdasarkan ID dengan middleware untuk otentikasi
	router.Handle("/inventory-id", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			controllers.GetInventoryByID(w, r) // Mengambil data inventaris berdasarkan ID
		case http.MethodPut:
			controllers.UpdateInventory(w, r) // Mengupdate data inventaris berdasarkan ID
		case http.MethodDelete:
			controllers.DeleteInventory(w, r) // Menghapus data inventaris berdasarkan ID
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

	return router
}
