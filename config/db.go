package config

import (
	"context"
	"time"

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
