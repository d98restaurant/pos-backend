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
)

type PaymentHandler struct {
    db *database.MongoDB
}

func NewPaymentHandler(db *database.MongoDB) *PaymentHandler {
    return &PaymentHandler{db: db}
}

// ProcessPayment processes a payment for an order
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
    orderID := c.Param("id")

    userID, _, _, ok := middleware.GetUserFromContext(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, models.APIResponse{
            Success: false,
            Message: "Authentication required",
            Error:   stringPtr("auth_required"),
        })
        return
    }

    var req models.ProcessPaymentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid request body",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    validMethods := map[string]bool{"cash": true, "credit_card": true, "debit_card": true, "digital_wallet": true}
    if !validMethods[req.PaymentMethod] {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid payment method",
            Error:   stringPtr("invalid_payment_method"),
        })
        return
    }

    if req.Amount <= 0 {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Payment amount must be greater than zero",
            Error:   stringPtr("invalid_amount"),
        })
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    ordersCollection := h.db.GetCollection("orders")
    var order bson.M
    err := ordersCollection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
    if err != nil {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Order not found",
            Error:   stringPtr("order_not_found"),
        })
        return
    }

    orderStatus, _ := order["status"].(string)
    if orderStatus == "cancelled" || orderStatus == "completed" {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Order cannot be paid - order is " + orderStatus,
            Error:   stringPtr("invalid_order_status"),
        })
        return
    }

    orderTotalAmount, _ := order["total_amount"].(float64)

    paymentsCollection := h.db.GetCollection("payments")
    cursor, _ := paymentsCollection.Find(ctx, bson.M{"order_id": orderID, "status": "completed"})
    var payments []bson.M
    cursor.All(ctx, &payments)
    totalPaid := 0.0
    for _, p := range payments {
        if amt, ok := p["amount"].(float64); ok {
            totalPaid += amt
        }
    }

    if totalPaid >= orderTotalAmount {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Order is already fully paid",
            Error:   stringPtr("order_fully_paid"),
        })
        return
    }

    remainingAmount := orderTotalAmount - totalPaid
    if req.Amount > remainingAmount {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Payment amount exceeds remaining balance",
            Error:   stringPtr("amount_exceeds_balance"),
        })
        return
    }

    paymentID := primitive.NewObjectID()
    now := time.Now()

    payment := bson.M{
        "_id":             paymentID,
        "order_id":        orderID,
        "payment_method":  req.PaymentMethod,
        "amount":          req.Amount,
        "reference_number": req.ReferenceNumber,
        "status":          "completed",
        "processed_by":    userID,
        "processed_at":    now,
        "created_at":      now,
    }

    _, err = paymentsCollection.InsertOne(ctx, payment)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to create payment record",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    newTotalPaid := totalPaid + req.Amount
    if newTotalPaid >= orderTotalAmount {
        ordersCollection.UpdateOne(ctx,
            bson.M{"_id": orderID},
            bson.M{"$set": bson.M{"status": "completed", "completed_at": now, "updated_at": now}},
        )

        if tableID, ok := order["table_id"].(string); ok && tableID != "" {
            tablesCollection := h.db.GetCollection("tables")
            tablesCollection.UpdateOne(ctx,
                bson.M{"_id": tableID},
                bson.M{"$set": bson.M{"is_occupied": false, "updated_at": now}},
            )
        }
    }

    c.JSON(http.StatusCreated, models.APIResponse{
        Success: true,
        Message: "Payment processed successfully",
        Data:    payment,
    })
}

// GetPayments retrieves payments for an order
func (h *PaymentHandler) GetPayments(c *gin.Context) {
    orderID := c.Param("id")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    ordersCollection := h.db.GetCollection("orders")
    var order bson.M
    if err := ordersCollection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order); err != nil {
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

    var payments []bson.M
    if err = cursor.All(ctx, &payments); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to scan payment",
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

// GetPaymentSummary retrieves payment summary for an order
func (h *PaymentHandler) GetPaymentSummary(c *gin.Context) {
    orderID := c.Param("id")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    ordersCollection := h.db.GetCollection("orders")
    var order bson.M
    err := ordersCollection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
    if err != nil {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Order not found",
            Error:   stringPtr("order_not_found"),
        })
        return
    }

    totalAmount, _ := order["total_amount"].(float64)

    paymentsCollection := h.db.GetCollection("payments")
    cursor, _ := paymentsCollection.Find(ctx, bson.M{"order_id": orderID, "status": "completed"})
    var payments []bson.M
    cursor.All(ctx, &payments)

    totalPaid := 0.0
    for _, p := range payments {
        if amt, ok := p["amount"].(float64); ok {
            totalPaid += amt
        }
    }

    remainingAmount := totalAmount - totalPaid
    summary := map[string]interface{}{
        "order_id":         orderID,
        "total_amount":     totalAmount,
        "total_paid":       totalPaid,
        "pending_amount":   0.0,
        "remaining_amount": remainingAmount,
        "is_fully_paid":    remainingAmount <= 0,
        "payment_count":    len(payments),
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Payment summary retrieved successfully",
        Data:    summary,
    })
}