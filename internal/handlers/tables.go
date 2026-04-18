package handlers

import (
    "context"
    "net/http"
    "time"

    "pos-backend/internal/database"
    "pos-backend/internal/models"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
)

type TableHandler struct {
    db *database.MongoDB
}

func NewTableHandler(db *database.MongoDB) *TableHandler {
    return &TableHandler{db: db}
}

func (h *TableHandler) GetTables(c *gin.Context) {
    collection := h.db.GetCollection("tables")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to fetch tables",
            Error:   stringPtr(err.Error()),
        })
        return
    }
    defer cursor.Close(ctx)

    var tables []models.MongoTable
    if err = cursor.All(ctx, &tables); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to parse tables",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Tables retrieved successfully",
        Data:    tables,
    })
}

func (h *TableHandler) GetTable(c *gin.Context) {
    tableID := c.Param("id")
    
    collection := h.db.GetCollection("tables")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var table models.MongoTable
    err := collection.FindOne(ctx, bson.M{"table_id": tableID}).Decode(&table)
    
    if err != nil {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Table not found",
            Error:   stringPtr("table_not_found"),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Table retrieved successfully",
        Data:    table,
    })
}
