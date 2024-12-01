package main

import (
    "fmt"
    "laundry-pos/config"
    "laundry-pos/middleware"
    "laundry-pos/routes"
    "net/http"
    "log"
)

func init() {
    // Menginisialisasi koneksi ke Supabase (PostgreSQL) melalui GORM
    if err := config.InitDB(); err != nil {
        fmt.Printf("Failed to initialize PostgreSQL: %v\n", err)
        panic(err)
    }
    fmt.Println("PostgreSQL initialized successfully!")
}

// Handler adalah fungsi yang akan dipanggil oleh server
func Handler(w http.ResponseWriter, r *http.Request) {
    // Inisialisasi router
    router := routes.InitRoutes()

    // Menambahkan middleware CORS
    routerWithCORS := middleware.EnableCORS(router)

    // Jalankan request melalui router
    routerWithCORS.ServeHTTP(w, r)
}

// Fungsi main untuk menjalankan server di lokal
func main() {
    http.HandleFunc("/", Handler)
    port := "8082"
    fmt.Printf("Server berjalan di http://localhost:%s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
