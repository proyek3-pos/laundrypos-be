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
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`    // ID menggunakan ObjectID untuk MongoDB
	FirstName   string             `json:"firstName" bson:"firstName"`
	LastName    string             `json:"lastName" bson:"lastName"`
	PhoneNumber string             `json:"phoneNumber" bson:"phoneNumber"`
}

// Model untuk Report (Laporan transaksi)
type Report struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"` // ID menggunakan ObjectID untuk MongoDB
	ReportDate        time.Time          `json:"reportDate" bson:"reportDate"`
	TotalTransactions int                `json:"totalTransactions" bson:"totalTransactions"`
	TotalIncome       float64            `json:"totalIncome" bson:"totalIncome"`
	TotalExpenses     float64            `json:"totalExpenses" bson:"totalExpenses"`
	NetProfit         float64            `json:"netProfit" bson:"netProfit"`
}

// Model untuk Inventory (Stok barang)
type Service struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`    // ID menggunakan ObjectID untuk MongoDB
	ServiceName string             `json:"serviceName" bson:"serviceName"` // Nama layanan (misalnya, "Cuci Kering")
	Description string             `json:"description" bson:"description"` // Deskripsi layanan
	UnitPrice   float64            `json:"unitPrice" bson:"unitPrice"`    // Harga per unit (misalnya, per kg)
}

// Model untuk Transaction (Transaksi)
type Transaction struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`        // ID menggunakan ObjectID untuk MongoDB
	TransactionID   string             `json:"transactionId" bson:"transactionId"` // ID Transaksi
	CustomerID      primitive.ObjectID `json:"customerId" bson:"customerId"` // Referensi ke Customer
	Customer        Customer           `json:"customer" bson:"customer"` // Embedded Customer Document (bisa menggunakan ID atau seluruh dokumen)
	TransactionDate time.Time          `json:"transactionDate" bson:"transactionDate"`
	Items           []TransactionItem  `json:"items" bson:"items"` // Item terkait transaksi
	TotalAmount     float64            `json:"totalAmount" bson:"totalAmount"`
	PaymentMethod   string             `json:"paymentMethod" bson:"paymentMethod"`
	Status          string             `json:"status" bson:"status"`
}

// Model untuk Item dalam transaksi
type TransactionItem struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"` // ID menggunakan ObjectID untuk MongoDB
	TransactionID primitive.ObjectID `json:"transactionId" bson:"transactionId"` // Referensi ke Transaksi
	Transaction   Transaction       `json:"transaction" bson:"transaction"`  // Embedded Transaction Document
	ServiceID    primitive.ObjectID `json:"serviceId" bson:"serviceId"`     // Referensi ke Service
	Service      Service             `json:"service" bson:"service"` // Embedded Service Document
	Quantity     int                 `json:"quantity" bson:"quantity"`
	UnitPrice    float64             `json:"unitPrice" bson:"unitPrice"`
	TotalPrice   float64             `json:"totalPrice" bson:"totalPrice"`
}

type Payment struct {
    OrderID     string          `json:"order_id" bson:"order_id"`
    GrossAmount float64         `json:"gross_amount" bson:"gross_amount"`
    SnapURL     string          `json:"snap_url" bson:"snap_url"`
    Customer    PaymentCustomer `json:"customer" bson:"customer"`
    Status      string          `json:"status" bson:"status"` // Status pembayaran (misalnya, Pending, Completed)
}

// Model untuk informasi Customer dalam pembayaran
type PaymentCustomer struct {
    Name  string `json:"name" bson:"name"`
    Email string `json:"email" bson:"email"`
    Phone string `json:"phone" bson:"phone"`
}