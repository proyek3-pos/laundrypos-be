package models

import (
	"time"

	"github.com/google/uuid" // Untuk UUID
)

// Model untuk User
type User struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Username string    `json:"username" gorm:"unique;not null"`
	Password string    `json:"password" gorm:"not null"`
	Role     string    `json:"role" gorm:"not null"` // Contoh: "admin" atau "staff"
}

// Model untuk Customer
type Customer struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	FirstName   string    `json:"firstName" gorm:"not null"`
	LastName    string    `json:"lastName" gorm:"not null"`
	PhoneNumber string    `json:"phoneNumber" gorm:"unique;not null"`
	Email       string    `json:"email" gorm:"unique;not null"`
	Address     string    `json:"address" gorm:"not null"`
}

// Model untuk Inventory (Stok barang)
type Service struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ServiceName string    `json:"serviceName" gorm:"not null"` // Nama layanan (misalnya, "Cuci Kering")
	Description string    `json:"description"`                 // Deskripsi layanan
	UnitPrice   float64   `json:"unitPrice" gorm:"not null"`   // Harga per unit (misalnya, per kg)
}

// Model untuk Report (Laporan transaksi)
type Report struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ReportDate        time.Time `json:"reportDate" gorm:"not null"`
	TotalTransactions int       `json:"totalTransactions" gorm:"not null"`
	TotalIncome       float64   `json:"totalIncome" gorm:"not null"`
	TotalExpenses     float64   `json:"totalExpenses" gorm:"not null"`
	NetProfit         float64   `json:"netProfit" gorm:"not null"`
}

// Model untuk Transaction (Transaksi)
type Transaction struct {
	ID              uuid.UUID         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TransactionID   string            `json:"transactionId" gorm:"unique;not null"`
	CustomerID      uuid.UUID         `json:"customerId" gorm:"not null"`
	Customer        Customer          `gorm:"foreignKey:CustomerID;references:ID"`
	TransactionDate time.Time         `json:"transactionDate" gorm:"not null"`
	Items           []TransactionItem `json:"items" gorm:"foreignKey:TransactionID;references:ID"`
	TotalAmount     float64           `json:"totalAmount" gorm:"not null"`
	PaymentMethod   string            `json:"paymentMethod" gorm:"not null"`
	Status          string            `json:"status" gorm:"not null"`
}

// Model untuk Item dalam transaksi
type TransactionItem struct {
    ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    TransactionID uuid.UUID `json:"transactionId" gorm:"not null"`
    Transaction   Transaction `gorm:"foreignKey:TransactionID;references:ID"`
    ServiceID    uuid.UUID `json:"serviceId" gorm:"not null"` // Mengarah ke Service
    Service      Service   `gorm:"foreignKey:ServiceID;references:ID"`
    Quantity     int       `json:"quantity" gorm:"not null"`
    UnitPrice    float64   `json:"unitPrice" gorm:"not null"`
    TotalPrice   float64   `json:"totalPrice" gorm:"not null"`
}
