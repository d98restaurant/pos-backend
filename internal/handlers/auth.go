package handlers

import (
    "context"
    "net/http"
    "time"

    "pos-backend/internal/database"
    "pos-backend/internal/middleware"
    "pos-backend/internal/models"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.mongodb.org/mongo-driver/bson"
    "golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
    db *database.MongoDB
}

func NewAuthHandler(db *database.MongoDB) *AuthHandler {
    return &AuthHandler{db: db}
}

// Login handles user authentication
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

    if req.Username == "" || req.Password == "" {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Username and password are required",
            Error:   stringPtr("missing_credentials"),
        })
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("users")
    var user bson.M
    err := collection.FindOne(ctx, bson.M{"username": req.Username, "is_active": true}).Decode(&user)

    if err != nil {
        c.JSON(http.StatusUnauthorized, models.APIResponse{
            Success: false,
            Message: "Invalid username or password",
            Error:   stringPtr("invalid_credentials"),
        })
        return
    }

    passwordHash, ok := user["password_hash"].(string)
    if !ok {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Database error",
            Error:   stringPtr("invalid_password_hash"),
        })
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, models.APIResponse{
            Success: false,
            Message: "Invalid username or password",
            Error:   stringPtr("invalid_credentials"),
        })
        return
    }

    userID, _ := uuid.Parse(user["_id"].(string))
    responseUser := models.User{
        ID:        userID,
        Username:  user["username"].(string),
        Email:     user["email"].(string),
        FirstName: user["first_name"].(string),
        LastName:  user["last_name"].(string),
        Role:      user["role"].(string),
        IsActive:  user["is_active"].(bool),
        CreatedAt: user["created_at"].(time.Time),
        UpdatedAt: user["updated_at"].(time.Time),
    }

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

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
    userID, _, _, ok := middleware.GetUserFromContext(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, models.APIResponse{
            Success: false,
            Message: "Authentication required",
            Error:   stringPtr("auth_required"),
        })
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("users")
    var user bson.M
    err := collection.FindOne(ctx, bson.M{"_id": userID.String()}).Decode(&user)

    if err != nil {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "User not found",
            Error:   stringPtr("user_not_found"),
        })
        return
    }

    responseUser := models.User{
        ID:        userID,
        Username:  user["username"].(string),
        Email:     user["email"].(string),
        FirstName: user["first_name"].(string),
        LastName:  user["last_name"].(string),
        Role:      user["role"].(string),
        IsActive:  user["is_active"].(bool),
        CreatedAt: user["created_at"].(time.Time),
        UpdatedAt: user["updated_at"].(time.Time),
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "User retrieved successfully",
        Data:    responseUser,
    })
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Logout successful",
    })
}