package handlers

import (
    "context"
    "net/http"
    "time"

    "pos-backend/internal/database"
    "pos-backend/internal/models"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.mongodb.org/mongo-driver/bson"
)

type PaymentHandler struct {
    db *database.MongoDB
}

func NewPaymentHandler(db *database.MongoDB) *PaymentHandler {
    return &PaymentHandler{db: db}
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
    orderID := c.Param("id")
    
    var req models.ProcessPaymentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid request body",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    paymentID := uuid.New().String()
    now := time.Now()
    
    payment := models.MongoPayment{
        PaymentID:     paymentID,
        OrderID:       orderID,
        PaymentMethod: req.PaymentMethod,
        Amount:        req.Amount,
        Status:        "completed",
        ProcessedAt:   &now,
        CreatedAt:     now,
    }
    
    if req.ReferenceNumber != nil {
        payment.ReferenceNumber = *req.ReferenceNumber
    }

    collection := h.db.GetCollection("payments")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err := collection.InsertOne(ctx, payment)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to process payment",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    // Update order status to completed
    ordersCollection := h.db.GetCollection("orders")
    _, err = ordersCollection.UpdateOne(
        ctx,
        bson.M{"order_id": orderID},
        bson.M{"$set": bson.M{"status": "completed", "completed_at": now}},
    )
    if err != nil {
        // Log error but don't fail the payment
        print("Failed to update order status: ", err.Error())
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Payment processed successfully",
        Data:    payment,
    })
}

func (h *PaymentHandler) GetPayments(c *gin.Context) {
    orderID := c.Param("id")
    
    collection := h.db.GetCollection("payments")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cursor, err := collection.Find(ctx, bson.M{"order_id": orderID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to fetch payments",
            Error:   stringPtr(err.Error()),
        })
        return
    }
    defer cursor.Close(ctx)

    var payments []models.MongoPayment
    if err = cursor.All(ctx, &payments); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to parse payments",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Payments retrieved successfully",
        Data:    payments,
    })
}

func (h *PaymentHandler) GetPaymentSummary(c *gin.Context) {
    orderID := c.Param("id")
    
    ordersCollection := h.db.GetCollection("orders")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var order models.MongoOrder
    err := ordersCollection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&order)
    
    if err != nil {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Order not found",
            Error:   stringPtr("order_not_found"),
        })
        return
    }

    paymentsCollection := h.db.GetCollection("payments")
    cursor, err := paymentsCollection.Find(ctx, bson.M{"order_id": orderID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to fetch payments",
            Error:   stringPtr(err.Error()),
        })
        return
    }
    defer cursor.Close(ctx)

    var payments []models.MongoPayment
    cursor.All(ctx, &payments)
    
    totalPaid := 0.0
    for _, p := range payments {
        totalPaid += p.Amount
    }

    summary := map[string]interface{}{
        "order_id":         orderID,
        "total_amount":     order.TotalAmount,
        "total_paid":       totalPaid,
        "pending_amount":   order.TotalAmount - totalPaid,
        "remaining_amount": order.TotalAmount - totalPaid,
        "is_fully_paid":    totalPaid >= order.TotalAmount,
        "payment_count":    len(payments),
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Payment summary retrieved successfully",
        Data:    summary,
    })
}
