package main

import (
	"fmt"
	"laundry-pos/config"
	"laundry-pos/middleware"
	"laundry-pos/routes"
	"log"
	"net/http"
)

func main() {
	// Menginisialisasi MongoDB
	if err := config.InitMongoDB(); err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	fmt.Println("MongoDB initialized successfully!")

	// Inisialisasi router
	router := routes.InitRoutes()

	// Menambahkan middleware CORS
	routerWithCORS := middleware.EnableCORS(router)

	// Jalankan server
	port := "8082"
	fmt.Printf("Server berjalan di http://localhost:%s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, routerWithCORS))
}
