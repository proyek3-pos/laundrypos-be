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
			controllers.Register(w, r) // Untuk registrasi staff
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

	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.Logout(w, r) // Memanggil fungsi logout
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Rute untuk customer
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


	// Rute untuk Service
    router.Handle("/services", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.CreateService(w, r) // Tambah layanan baru
		case http.MethodGet:
			controllers.GetAllServices(w, r) // Ambil semua data layanan
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

    router.Handle("/service-id", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			controllers.GetServiceByID(w, r) // Ambil layanan berdasarkan ID
		case http.MethodPut:
			controllers.UpdateService(w, r) // Update layanan
		case http.MethodDelete:
			controllers.DeleteService(w, r) // Hapus layanan
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Rute untuk Transaksi
    router.Handle("/transactions", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.CreateTransaction(w, r) // Buat transaksi baru
		case http.MethodGet:
			controllers.GetTransactions(w, r) // Ambil semua transaksi
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

    router.Handle("/transaction-id", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			controllers.GetTransactionByID(w, r) // Ambil transaksi berdasarkan ID
		case http.MethodPut:
			controllers.UpdateTransaction(w, r) // Update layanan
		case http.MethodDelete:
			controllers.DeleteTransaction(w, r) // Hapus layanan
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Rute untuk membuat pembayaran menggunakan Midtrans
	router.HandleFunc("/create-payment", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.CreatePayment(w, r) // Memanggil fungsi CreatePayment untuk membuat pembayaran
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Rute untuk pembayaran
	router.Handle("/payments", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			controllers.GetAllPayments(w, r) // Mengambil semua pembayaran
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

	router.Handle("/payment-detail", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			controllers.GetPaymentByOrderID(w, r) // Mengambil pembayaran berdasarkan OrderID
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))


	// Tambahkan handler untuk menerima webhook dari Midtrans
	router.HandleFunc("/webhook/midtrans", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.WebhookHandler(w, r) // Memanggil fungsi WebhookHandler untuk menangani notifikasi webhook
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	return router
}
