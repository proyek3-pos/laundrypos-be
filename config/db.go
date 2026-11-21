package config

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var UserCollection *mongo.Collection
var CustomerCollection *mongo.Collection
var PaymentCollection *mongo.Collection
var ServiceCollection *mongo.Collection
var TransactionCollection *mongo.Collection
var RevokedTokenCollection *mongo.Collection

func ConnectMongo(cfg Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	Client = client
	db := client.Database(cfg.DBName)
	UserCollection = db.Collection("user")
	CustomerCollection = db.Collection("customers")
	PaymentCollection = db.Collection("payments")
	ServiceCollection = db.Collection("service")
	TransactionCollection = db.Collection("transactions")
	RevokedTokenCollection = db.Collection("revoked_tokens")
	return nil
}

// EnsureIndexes membuat index yang sistem andalkan: username unik (anti race
// register), order_id unik, dan TTL untuk token yang dicabut.
func EnsureIndexes(ctx context.Context) error {
	_, err := UserCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"username": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	_, err = PaymentCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"order_id": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	// Satu tagihan Midtrans pending per transaksi; menahan race dobel Snap.
	_, err = PaymentCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"transactionId": 1},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"payment_method": "midtrans", "status": "pending"}),
	})
	if err != nil {
		return err
	}
	// Nomor nota harus unik, tapi hanya untuk dokumen yang punya nota;
	// transaksi lama sebelum fitur nota tidak membawa field ini.
	_, err = TransactionCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"invoiceNumber": 1},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{
			"invoiceNumber": bson.M{"$type": "string", "$gt": ""},
		}),
	})
	if err != nil {
		return err
	}
	_, err = TransactionCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.M{"customerId": 1}},
		{Keys: bson.M{"status": 1}},
		{Keys: bson.M{"transactionDate": -1}},
	})
	if err != nil {
		return err
	}
	_, err = RevokedTokenCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"expiresAt": 1},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	return err
}
