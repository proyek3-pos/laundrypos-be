package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Role     string `json:"role" bson:"role"` // Bisa berisi nilai seperti "admin" atau "staff"
}

// Model untuk Customer
type Customer struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FirstName   string             `bson:"firstName" json:"firstName"`
	LastName    string             `bson:"lastName" json:"lastName"`
	PhoneNumber string             `bson:"phoneNumber" json:"phoneNumber"`
	Email       string             `bson:"email" json:"email"`
	Address     string             `bson:"address" json:"address"`
}

// Model untuk Inventory (Stok barang)
type Inventory struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ItemName     string             `bson:"itemName" json:"itemName"`
	Quantity     int                `bson:"quantity" json:"quantity"`
	UnitPrice    float64            `bson:"unitPrice" json:"unitPrice"`
	Supplier     string             `bson:"supplier" json:"supplier"`
	Expiration   time.Time          `bson:"expiration" json:"expiration"`
	ReorderLevel int                `bson:"reorderLevel" json:"reorderLevel"`
}

// Model untuk Report (Laporan transaksi)
type Report struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ReportDate        time.Time          `bson:"reportDate" json:"reportDate"`
	TotalTransactions int                `bson:"totalTransactions" json:"totalTransactions"`
	TotalIncome       float64            `bson:"totalIncome" json:"totalIncome"`
	TotalExpenses     float64            `bson:"totalExpenses" json:"totalExpenses"`
	NetProfit         float64            `bson:"netProfit" json:"netProfit"`
}

// Model untuk Transaction (Transaksi)
type Transaction struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TransactionID   string             `bson:"transactionId" json:"transactionId"`
	CustomerID      primitive.ObjectID `bson:"customerId" json:"customerId"`
	TransactionDate time.Time          `bson:"transactionDate" json:"transactionDate"`
	Items           []TransactionItem  `bson:"items" json:"items"`
	TotalAmount     float64            `bson:"totalAmount" json:"totalAmount"`
	PaymentMethod   string             `bson:"paymentMethod" json:"paymentMethod"`
	Status          string             `bson:"status" json:"status"` // e.g., 'paid', 'pending'
}

// Submodel untuk Item dalam transaksi
type TransactionItem struct {
	ItemID     primitive.ObjectID `bson:"itemId" json:"itemId"`
	ItemName   string             `bson:"itemName" json:"itemName"`
	Quantity   int                `bson:"quantity" json:"quantity"`
	UnitPrice  float64            `bson:"unitPrice" json:"unitPrice"`
	TotalPrice float64            `bson:"totalPrice" json:"totalPrice"`
}