package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username string             `json:"username" bson:"username"`
	Password string             `json:"password" bson:"password"`
	Role     string             `json:"role" bson:"role"` // admin | staff
}

type Customer struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FullName    string             `json:"fullName" bson:"fullName"`
	Email       string             `json:"email" bson:"email"`
	PhoneNumber string             `json:"phoneNumber" bson:"phoneNumber"`
}

type Service struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ServiceName string             `json:"serviceName" bson:"serviceName"`
	Description string             `json:"description" bson:"description"`
	UnitPrice   float64            `json:"unitPrice" bson:"unitPrice"`
	Unit        string             `json:"unit" bson:"unit"` // kg | item
}

type StatusEvent struct {
	Status string    `json:"status" bson:"status"`
	At     time.Time `json:"at" bson:"at"`
	By     string    `json:"by" bson:"by"`
}

type Transaction struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	InvoiceNumber   string             `json:"invoiceNumber" bson:"invoiceNumber"`
	CustomerID      primitive.ObjectID `json:"customerId" bson:"customerId"`
	Customer        Customer           `json:"customer" bson:"customer,omitempty"`
	TransactionDate time.Time          `json:"transactionDate" bson:"transactionDate"`
	Items           []TransactionItem  `json:"items" bson:"items"`
	TotalAmount     float64            `json:"totalAmount" bson:"totalAmount"`
	PaymentMethod   string             `json:"paymentMethod" bson:"paymentMethod,omitempty"` // cash | midtrans
	PaymentStatus   string             `json:"paymentStatus" bson:"paymentStatus"`           // unpaid | paid
	PaidAt          time.Time          `json:"paidAt,omitempty" bson:"paidAt,omitempty"`
	Status          string             `json:"status" bson:"status"` // pending | diproses | selesai | diambil | batal
	StatusHistory   []StatusEvent      `json:"statusHistory" bson:"statusHistory"`
}

type TransactionItem struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ServiceID  primitive.ObjectID `json:"serviceId" bson:"serviceId"`
	Service    Service            `json:"service" bson:"service"`
	Quantity   float64            `json:"quantity" bson:"quantity"`
	UnitPrice  float64            `json:"unitPrice" bson:"unitPrice"`
	TotalPrice float64            `json:"totalPrice" bson:"totalPrice"`
}

type Payment struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TransactionID primitive.ObjectID `json:"transactionId" bson:"transactionId"`
	OrderID       string             `json:"order_id" bson:"order_id"`
	GrossAmount   float64            `json:"gross_amount" bson:"gross_amount"`
	AmountPaid    float64            `json:"amount_paid,omitempty" bson:"amount_paid,omitempty"` // hanya cash
	Change        float64            `json:"change,omitempty" bson:"change,omitempty"`           // kembalian cash
	SnapURL       string             `json:"snap_url,omitempty" bson:"snap_url,omitempty"`
	Status        string             `json:"status" bson:"status"` // pending | paid | failed
	PaymentMethod string             `json:"payment_method,omitempty" bson:"payment_method,omitempty"`
	MidtransRaw   string             `json:"midtrans_raw,omitempty" bson:"midtrans_raw,omitempty"` // status mentah terakhir dari Midtrans
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	PaidAt        time.Time          `json:"paid_at,omitempty" bson:"paid_at,omitempty"`
}
