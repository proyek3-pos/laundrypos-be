package routes

import (
	"laundry-pos/controllers"
	"laundry-pos/middleware"
	"net/http"
)

func InitRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	auth := middleware.AuthMiddleware
	admin := func(next http.Handler) http.Handler {
		return auth(middleware.RequireRole("admin", next))
	}
	h := func(next http.HandlerFunc) http.Handler {
		return auth(next)
	}

	mux.HandleFunc("POST /register", controllers.Register)
	mux.HandleFunc("POST /login", controllers.Login)
	mux.HandleFunc("GET /health", controllers.Health)
	mux.Handle("POST /logout", auth(http.HandlerFunc(controllers.Logout)))

	mux.Handle("POST /customers", h(controllers.AddCustomer))
	mux.Handle("GET /customers", h(controllers.GetAllCustomers))
	mux.Handle("GET /customers/{id}", h(controllers.GetCustomerByID))
	mux.Handle("PUT /customers/{id}", h(controllers.UpdateCustomer))
	mux.Handle("DELETE /customers/{id}", h(controllers.DeleteCustomer))

	mux.Handle("GET /services", h(controllers.GetAllServices))
	mux.Handle("POST /services", admin(http.HandlerFunc(controllers.CreateService)))
	mux.Handle("GET /services/{id}", h(controllers.GetServiceByID))
	mux.Handle("PUT /services/{id}", admin(http.HandlerFunc(controllers.UpdateService)))
	mux.Handle("DELETE /services/{id}", admin(http.HandlerFunc(controllers.DeleteService)))

	mux.Handle("POST /transactions", h(controllers.CreateTransaction))
	mux.Handle("GET /transactions", h(controllers.GetTransactions))
	mux.Handle("GET /transactions/{id}", h(controllers.GetTransactionByID))
	mux.Handle("PUT /transactions/{id}", h(controllers.UpdateTransaction))
	mux.Handle("PATCH /transactions/{id}/status", h(controllers.UpdateTransactionStatus))
	mux.Handle("DELETE /transactions/{id}", admin(http.HandlerFunc(controllers.DeleteTransaction)))

	mux.Handle("POST /payments/midtrans", h(controllers.CreatePayment))
	mux.Handle("POST /payments/cash", h(controllers.CreateCashPayment))
	mux.Handle("GET /payments", h(controllers.GetAllPayments))
	mux.Handle("GET /payments/{orderID}", h(controllers.GetPaymentByOrderID))

	mux.Handle("GET /reports/summary", admin(http.HandlerFunc(controllers.GetReportsSummary)))

	// Webhook Midtrans memang harus publik, tapi diverifikasi signature.
	mux.HandleFunc("POST /webhook/midtrans", controllers.WebhookHandler)

	return mux
}
