package handlers

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "pos-backend/internal/database"
    "pos-backend/internal/models"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.mongodb.org/mongo-driver/bson"
)

type OrderHandler struct {
    db *database.MongoDB
}

func NewOrderHandler(db *database.MongoDB) *OrderHandler {
    return &OrderHandler{db: db}
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
    collection := h.db.GetCollection("orders")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to fetch orders",
            Error:   stringPtr(err.Error()),
        })
        return
    }
    defer cursor.Close(ctx)

    var orders []models.MongoOrder
    if err = cursor.All(ctx, &orders); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to parse orders",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Orders retrieved successfully",
        Data:    orders,
    })
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
    orderID := c.Param("id")
    
    collection := h.db.GetCollection("orders")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var order models.MongoOrder
    err := collection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&order)
    
    if err != nil {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Order not found",
            Error:   stringPtr("order_not_found"),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Order retrieved successfully",
        Data:    order,
    })
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
    var req models.CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid request body",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    orderID := uuid.New().String()
    orderNumber := fmt.Sprintf("ORD%d", time.Now().UnixNano())

    order := models.MongoOrder{
        OrderID:     orderID,
        OrderNumber: orderNumber,
        OrderType:   req.OrderType,
        Status:      "pending",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    if req.TableID != nil {
        order.TableID = *req.TableID
    }
    if req.CustomerName != nil {
        order.CustomerName = *req.CustomerName
    }
    if req.Notes != nil {
        order.Notes = *req.Notes
    }

    collection := h.db.GetCollection("orders")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err := collection.InsertOne(ctx, order)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to create order",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    c.JSON(http.StatusCreated, models.APIResponse{
        Success: true,
        Message: "Order created successfully",
        Data:    order,
    })
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
    orderID := c.Param("id")
    
    var req struct {
        Status string `json:"status"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid request body",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    collection := h.db.GetCollection("orders")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    update := bson.M{
        "$set": bson.M{
            "status":     req.Status,
            "updated_at": time.Now(),
        },
    }

    result, err := collection.UpdateOne(ctx, bson.M{"order_id": orderID}, update)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to update order status",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Order not found",
            Error:   stringPtr("order_not_found"),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Order status updated successfully",
    })
}

func (h *OrderHandler) CreateDineInOrder(c *gin.Context) {
    var req models.CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid request body",
            Error:   stringPtr(err.Error()),
        })
        return
    }
    
    req.OrderType = "dine_in"
    h.CreateOrder(c)
}
