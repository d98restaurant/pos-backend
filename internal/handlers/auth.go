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

    collection := h.db.GetCollection("users")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

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

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, models.APIResponse{
            Success: false,
            Message: "Invalid username or password",
            Error:   stringPtr("invalid_credentials"),
        })
        return
    }

    // Convert ObjectID to UUID string
    userUUID, _ := uuid.Parse(user.UserID)
    if userUUID == uuid.Nil {
        userUUID = uuid.New()
    }

    responseUser := models.User{
        ID:        userUUID,
        Username:  user.Username,
        Email:     user.Email,
        FirstName: user.FirstName,
        LastName:  user.LastName,
        Role:      user.Role,
        IsActive:  user.IsActive,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
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

    collection := h.db.GetCollection("users")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var user models.MongoUser
    err := collection.FindOne(ctx, bson.M{"user_id": userID.String()}).Decode(&user)
    
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
        Username:  user.Username,
        Email:     user.Email,
        FirstName: user.FirstName,
        LastName:  user.LastName,
        Role:      user.Role,
        IsActive:  user.IsActive,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "User retrieved successfully",
        Data:    responseUser,
    })
}

func (h *AuthHandler) Logout(c *gin.Context) {
    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Logout successful",
    })
}
