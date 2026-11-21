package controllers

import (
	"context"
	"encoding/json"
	"laundry-pos/config"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func GetReportsSummary(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	revenueToday, err := sumPaidRevenue(ctx, startOfDay)
	if err != nil {
		http.Error(w, "Gagal menghitung pendapatan harian", http.StatusInternalServerError)
		return
	}
	revenueMonth, err := sumPaidRevenue(ctx, startOfMonth)
	if err != nil {
		http.Error(w, "Gagal menghitung pendapatan bulanan", http.StatusInternalServerError)
		return
	}
	ordersToday, err := config.TransactionCollection.CountDocuments(ctx, bson.M{
		"transactionDate": bson.M{"$gte": startOfDay},
		"status":          bson.M{"$ne": "batal"},
	})
	if err != nil {
		http.Error(w, "Gagal menghitung pesanan", http.StatusInternalServerError)
		return
	}
	unpaid, err := config.TransactionCollection.CountDocuments(ctx, bson.M{
		"paymentStatus": "unpaid",
		"status":        bson.M{"$ne": "batal"},
	})
	if err != nil {
		http.Error(w, "Gagal menghitung pesanan belum lunas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"revenueToday":   revenueToday,
		"revenueMonth":   revenueMonth,
		"ordersToday":    ordersToday,
		"unpaidCount":    unpaid,
	})
}

// sumPaidRevenue menjumlahkan transaksi lunas sejak waktu tertentu.
func sumPaidRevenue(ctx context.Context, since time.Time) (float64, error) {
	cursor, err := config.TransactionCollection.Aggregate(ctx, []bson.M{
		{"$match": bson.M{"paymentStatus": "paid", "paidAt": bson.M{"$gte": since}}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$totalAmount"}}},
	})
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		Total float64 `bson:"total"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].Total, nil
}
