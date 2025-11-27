package main

import (
    "log"
    "os"

    "backend-mobile-skill-test/config"
    "backend-mobile-skill-test/controllers"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
)

func main() {
    // 1. Load file .env
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }
    
    // Set mode Gin (opsional)
    // gin.SetMode(gin.ReleaseMode) 

    // 2. Koneksi ke Database
    config.ConnectPostgres()
    config.ConnectMongo() 
    
    // PENTING: Inisialisasi Collection MongoDB setelah klien terhubung
    controllers.InitProductCollection() 
    
    defer config.DB.Close()
    // Di lingkungan produksi, Anda mungkin ingin defer config.MongoCLient.Disconnect()

    // 3. Setup Router Gin
    router := gin.Default()

    // --- Definisi Endpoint Users (PostgreSQL) ---
    // Semua endpoint menggunakan method POST dan mengirim ID di body sesuai permintaan tes
    router.POST("/users/create", controllers.CreateUser)
    router.POST("/users/list", controllers.GetUserList)
    router.POST("/users/detail", controllers.GetUserByID)
    router.POST("/users/update", controllers.UpdateUser)
    router.POST("/users/delete", controllers.DeleteUser)

    // --- Definisi Endpoint Products (MongoDB) ---
    // Semua endpoint menggunakan method POST dan mengirim ID di body sesuai permintaan tes
    router.POST("/product/create", controllers.CreateProduct)
    router.POST("/product/list", controllers.GetProductList)
    router.POST("/product/detail", controllers.GetProductByID)
    router.POST("/product/update", controllers.UpdateProduct)
    router.POST("/product/delete", controllers.DeleteProduct)

    // 4. Jalankan Server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080" // Default port
    }
    log.Printf("Server running on port %s", port)
    if err := router.Run(":" + port); err != nil {
        log.Fatalf("Server failed to run: %v", err)
    }
}