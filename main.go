package main

import (
    "log"
    "os"

    "pos-backend/internal/api"
    "pos-backend/internal/database"
    "pos-backend/internal/middleware"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using environment variables")
    }

    // MongoDB connection
    mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
    mongoDBName := getEnv("MONGODB_DATABASE", "pos_system")

    db, err := database.ConnectMongoDB(mongoURI, mongoDBName)
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    defer db.Close()

    log.Println("Successfully connected to MongoDB")

    gin.SetMode(getEnv("GIN_MODE", "release"))
    router := gin.New()

    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    router.Use(cors.New(cors.Config{
        AllowOrigins: []string{
            "http://localhost:3000",
            "http://localhost:3001",
            "http://localhost:3002",
            "http://localhost:3003",
            "http://localhost:5173",
            "https://pos-system-c3c29.web.app",
            "https://pos-system-c3c29.firebaseapp.com",
        },
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With"},
        AllowCredentials: true,
    }))

    authMiddleware := middleware.AuthMiddleware()

    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy", "message": "POS API is running with MongoDB"})
    })

    apiRoutes := router.Group("/api/v1")
    api.SetupRoutes(apiRoutes, db, authMiddleware)

    port := getEnv("PORT", "8080")
    log.Printf("Starting server on port %s", port)

    if err := router.Run(":" + port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
