package main

import (
	"context"
	"fmt"
	"laundry-pos/config"
	"laundry-pos/services"
	"log"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Migrate menyetel data lama ke skema baru: status "Pending" -> "pending"
// dan paymentStatus default "unpaid". Jalankan sekali: go run ./cmd/migrate
func main() {
	_ = godotenv.Load()

	if err := config.UseLocalTZ(); err != nil {
		log.Fatal(err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err := config.ConnectMongo(cfg); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := config.TransactionCollection.UpdateMany(ctx,
		bson.M{"status": "Pending"},
		bson.M{"$set": bson.M{"status": "pending"}},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("status Pending -> pending: %d dokumen\n", res.ModifiedCount)

	res, err = config.TransactionCollection.UpdateMany(ctx,
		bson.M{"paymentStatus": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"paymentStatus": "unpaid"}},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("paymentStatus default: %d dokumen\n", res.ModifiedCount)

	// Isi nomor nota untuk transaksi lama yang belum punya, agar bisa difilter.
	cur, err := config.TransactionCollection.Find(ctx, bson.M{"invoiceNumber": bson.M{"$in": bson.A{nil, ""}}})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	var retro int
	for cur.Next(ctx) {
		var doc struct {
			ID              primitive.ObjectID `bson:"_id"`
			TransactionDate time.Time          `bson:"transactionDate"`
		}
		if err := cur.Decode(&doc); err != nil {
			continue
		}
		_, err := config.TransactionCollection.UpdateOne(ctx,
			bson.M{"_id": doc.ID},
			bson.M{"$set": bson.M{"invoiceNumber": services.NewInvoiceNumber(doc.TransactionDate)}},
		)
		if err == nil {
			retro++
		}
	}
	fmt.Printf("nomor nota retroaktif: %d dokumen\n", retro)
}
