package controllers

import (
	"context"
	"laundry-pos/config"
	"laundry-pos/utils"
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
		utils.Err(w, http.StatusInternalServerError, "Gagal menghitung pendapatan harian")
		return
	}
	revenueMonth, err := sumPaidRevenue(ctx, startOfMonth)
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal menghitung pendapatan bulanan")
		return
	}
	ordersToday, err := config.TransactionCollection.CountDocuments(ctx, bson.M{
		"transactionDate": bson.M{"$gte": startOfDay},
		"status":          bson.M{"$ne": "batal"},
	})
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal menghitung pesanan")
		return
	}
	unpaid, err := config.TransactionCollection.CountDocuments(ctx, bson.M{
		"paymentStatus": "unpaid",
		"status":        bson.M{"$ne": "batal"},
	})
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal menghitung pesanan belum lunas")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{
		"revenueToday": revenueToday,
		"revenueMonth": revenueMonth,
		"ordersToday":  ordersToday,
		"unpaidCount":  unpaid,
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
