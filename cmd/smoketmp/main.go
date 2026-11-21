package main

import (
	"context"
	"fmt"
	"laundry-pos/config"
	"log"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err := config.ConnectMongo(cfg); err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fmt.Println("dbname:", cfg.DBName)
	cur, _ := config.UserCollection.Find(ctx, bson.M{})
	defer cur.Close(ctx)
	var docs []bson.M
	if err := cur.All(ctx, &docs); err != nil {
		log.Fatal(err)
	}
	for _, d := range docs {
		if u, ok := d["username"].(string); ok {
			fmt.Println("-", u)
		}
	}
	fmt.Println("total:", len(docs))
}
