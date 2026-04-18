package handlers

import (
    "context"
    "net/http"
    "time"

    "pos-backend/internal/database"
    "pos-backend/internal/models"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "github.com/google/uuid"
)

type ProductHandler struct {
    db *database.MongoClient
}

func NewProductHandler(db *database.MongoClient) *ProductHandler {
    return &ProductHandler{db: db}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
    collection := h.db.GetCollection("products")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to fetch products",
            Error:   stringPtr(err.Error()),
        })
        return
    }
    defer cursor.Close(ctx)

    var products []models.MongoProduct
    if err = cursor.All(ctx, &products); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to parse products",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Products retrieved successfully",
        Data:    products,
    })
}