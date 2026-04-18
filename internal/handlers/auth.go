package handlers

import (
    "context"
    "net/http"
    "time"

    "pos-backend/internal/database"
    "pos-backend/internal/middleware"
    "pos-backend/internal/models"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
    db *database.MongoClient
}

func NewAuthHandler(db *database.MongoClient) *AuthHandler {
    return &AuthHandler{db: db}
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req models.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid request body",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    // Get users collection
    collection := h.db.GetCollection("users")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Find user by username
    var user models.MongoUser
    err := collection.FindOne(ctx, bson.M{"username": req.Username}).Decode(&user)
    
    if err != nil {
        c.JSON(http.StatusUnauthorized, models.APIResponse{
            Success: false,
            Message: "Invalid username or password",
            Error:   stringPtr("invalid_credentials"),
        })
        return
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, models.APIResponse{
            Success: false,
            Message: "Invalid username or password",
            Error:   stringPtr("invalid_credentials"),
        })
        return
    }

    // Convert to response format
    responseUser := models.User{
        ID:        user.ID.Hex(),
        Username:  user.Username,
        Email:     user.Email,
        FirstName: user.FirstName,
        LastName:  user.LastName,
        Role:      user.Role,
        IsActive:  user.IsActive,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }

    // Generate JWT token
    token, err := middleware.GenerateToken(&responseUser)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to generate token",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Login successful",
        Data: models.LoginResponse{
            Token: token,
            User:  responseUser,
        },
    })
}

func stringPtr(s string) *string {
    return &s
}