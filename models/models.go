package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Model untuk User
type User struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"` // ID menggunakan ObjectID untuk MongoDB
	Username string             `json:"username" bson:"username"`
	Password string             `json:"password" bson:"password"`
	Role     string             `json:"role" bson:"role"` // Contoh: "admin" atau "staff"
}

// Model untuk Customer
type Customer struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FullName    string             `json:"fullName" bson:"fullName"`       // Nama lengkap customer
	Email       string             `json:"email" bson:"email"`             // Email customer
	PhoneNumber string             `json:"phoneNumber" bson:"phoneNumber"` // Nomor telepon
}

// Model untuk Inventory (Stok barang)
type Service struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ServiceName string             `json:"serviceName" bson:"serviceName"`
	Description string             `json:"description" bson:"description"`
	UnitPrice   float64            `json:"unitPrice" bson:"unitPrice"`
	Unit        string             `json:"unit" bson:"unit"` // Misalnya, "kg", "item", dll.
}


// Model untuk Transaction (Transaksi)
type Transaction struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CustomerID      primitive.ObjectID `json:"customerId" bson:"customerId"`
	Customer        Customer           `json:"customer" bson:"customer,omitempty"` // Tambahkan ini untuk menyimpan informasi customer
	TransactionDate time.Time          `json:"transactionDate" bson:"transactionDate"`
	Items           []TransactionItem  `json:"items" bson:"items"`       // Daftar item dalam transaksi
	TotalAmount     float64            `json:"totalAmount" bson:"totalAmount"`
	PaymentMethod   string             `json:"paymentMethod" bson:"paymentMethod"`
	SnapURL         string             `json:"snap_url" bson:"snap_url"` // URL pembayaran Midtrans
	Status          string             `json:"status" bson:"status"`     // Status transaksi
}


// Model untuk Item dalam transaksi
type TransactionItem struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ServiceID primitive.ObjectID `json:"serviceId" bson:"serviceId"`
	Service   Service            `json:"service" bson:"service"` // Embedded Service Document
	Quantity  int                `json:"quantity" bson:"quantity"`
	UnitPrice float64            `json:"unitPrice" bson:"unitPrice"`
	TotalPrice float64           `json:"totalPrice" bson:"totalPrice"`
}


type Payment struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`       // ID pembayaran
	TransactionID primitive.ObjectID `json:"transactionId" bson:"transactionId"` // Referensi ke transaksi terkait
	OrderID       string             `json:"order_id" bson:"order_id"`      // ID pesanan unik untuk pembayaran
	GrossAmount   float64            `json:"gross_amount" bson:"gross_amount"` // Jumlah total yang harus dibayar
	SnapURL       string             `json:"snap_url" bson:"snap_url"`      // URL untuk pembayaran di Midtrans
	Status        string             `json:"status" bson:"status"`         // Status pembayaran: Pending, Success, Failed
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"` // Waktu pembuatan pembayaran
	PaymentMethod string             `json:"payment_method" bson:"payment_method,omitempty"` // Metode pembayaran jika diperlukan
}



// Model untuk informasi Customer dalam pembayaran
type PaymentCustomer struct {
    Name  string `json:"name" bson:"name"`
    Email string `json:"email" bson:"email"`
    Phone string `json:"phone" bson:"phone"`
}