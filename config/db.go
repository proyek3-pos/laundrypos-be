package config

import (
    "context"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var UserCollection *mongo.Collection
var CustomerCollection *mongo.Collection
var InventoryCollection *mongo.Collection
var PaymentCollection *mongo.Collection
var ServiceCollection *mongo.Collection
var TransactionCollection *mongo.Collection

// InitMongoDB untuk menginisialisasi koneksi ke MongoDB
func InitMongoDB() error {
    uri := "mongodb+srv://karamissuu:karamissu1@cluster0.lyovb.mongodb.net/?retryWrites=true&w=majority"
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    clientOptions := options.Client().ApplyURI(uri)
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal("Failed to connect to MongoDB: ", err)
        return err
    }

    if err := client.Ping(ctx, nil); err != nil {
        log.Fatal("MongoDB Ping failed: ", err)
        return err
    }

    log.Println("MongoDB connected successfully")
    Client = client
    UserCollection = Client.Database("laundry-pos").Collection("user")
    CustomerCollection = Client.Database("laundry-pos").Collection("customers")
    InventoryCollection = Client.Database("laundry-pos").Collection("inventory")
	PaymentCollection = client.Database("laundry-pos").Collection("payments")
	ServiceCollection = client.Database("laundry-pos").Collection("service")
	TransactionCollection = client.Database("laundry-pos").Collection("transactions")

    return nil
}
