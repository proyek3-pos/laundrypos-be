package main

import (
    "fmt"
    "log"
    "laundry-pos/config"
    "laundry-pos/routes"
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

    // Jalankan server
    port := "8082"
    fmt.Printf("Server berjalan di http://localhost:%s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, router))
}
