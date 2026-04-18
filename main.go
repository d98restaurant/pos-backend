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
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using environment variables")
    }

    // Connect to MongoDB Atlas
    mongoURI := getEnv("MONGODB_URI", "mongodb+srv://admin:mantu1996@cluster0.ap02ozs.mongodb.net/?appName=Cluster0")
    mongoDBName := getEnv("MONGODB_DATABASE", "pos_system")
    
    mongoClient, err := database.NewMongoClient(mongoURI, mongoDBName)
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    defer mongoClient.Close()

    log.Println("Successfully connected to MongoDB Atlas")

    // Initialize Gin router
    gin.SetMode(getEnv("GIN_MODE", "release"))
    router := gin.New()

    // Add middleware
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
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        AllowCredentials: true,
    }))

    // Add authentication middleware
    authMiddleware := middleware.AuthMiddleware()

    // Health check endpoint
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "healthy",
            "message": "POS API is running with MongoDB Atlas",
            "database": "MongoDB",
        })
    })

    // Initialize API routes (pass MongoDB client)
    apiRoutes := router.Group("/api/v1")
    api.SetupRoutes(apiRoutes, mongoClient, authMiddleware)

    // Start server
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