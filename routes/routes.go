package routes

import (
	"laundry-pos/controllers"
	"net/http"
)

func InitRoutes() *http.ServeMux {
	router := http.NewServeMux()

	// Rute Auth
	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.Register(w, r) // Untuk registrasi staff
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// router.HandleFunc("/admin/register", func(w http.ResponseWriter, r *http.Request) {
	// 	switch r.Method {
	// 	case http.MethodPost:
	// 		controllers.RegisterAdmin(w, r) // Untuk registrasi admin
	// 	default:
	// 		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	// 	}
	// })

	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.Login(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Rute untuk customer dengan middleware untuk otentikasi
	router.Handle("/customers", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.AddCustomer(w, r) // Membuat customer baru
		case http.MethodGet:
			controllers.GetAllCustomers(w, r) // Mengambil semua data customer
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Rute untuk mengambil customer berdasarkan ID dengan middleware untuk otentikasi
	router.Handle("/customer-id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))

	// // Rute untuk mencari atau membuat customer baru berdasarkan nama dan nomor telepon
	// router.Handle("/findcustomer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	switch r.Method {
	// 	case http.MethodPost:
	// 		controllers.FindOrCreateCustomer(w, r) // Memanggil fungsi FindOrCreateCustomer
	// 	default:
	// 		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	// 	}
	// }))

	// Rute untuk membuat pembayaran menggunakan Midtrans
	router.HandleFunc("/create-payment", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.CreatePayment(w, r) // Memanggil fungsi CreatePayment untuk membuat pembayaran
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	return router
}
