package handlers

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "pos-backend/internal/database"
    "pos-backend/internal/middleware"
    "pos-backend/internal/models"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type OrderHandler struct {
    db *database.MongoDB
}

func NewOrderHandler(db *database.MongoDB) *OrderHandler {
    return &OrderHandler{db: db}
}

// GetOrders retrieves all orders with pagination and filtering
func (h *OrderHandler) GetOrders(c *gin.Context) {
    page := 1
    perPage := 20
    status := c.Query("status")
    orderType := c.Query("order_type")

    if pageStr := c.Query("page"); pageStr != "" {
        fmt.Sscanf(pageStr, "%d", &page)
    }
    if perPageStr := c.Query("per_page"); perPageStr != "" {
        fmt.Sscanf(perPageStr, "%d", &perPage)
    }

    offset := int64((page - 1) * perPage)

    filter := bson.M{}
    if status != "" {
        filter["status"] = status
    }
    if orderType != "" {
        filter["order_type"] = orderType
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("orders")
    total, _ := collection.CountDocuments(ctx, filter)

    findOptions := options.Find()
    findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
    findOptions.SetLimit(int64(perPage))
    findOptions.SetSkip(offset)

    cursor, err := collection.Find(ctx, filter, findOptions)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to fetch orders",
            Error:   stringPtr(err.Error()),
        })
        return
    }
    defer cursor.Close(ctx)

    var orders []bson.M
    if err = cursor.All(ctx, &orders); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to parse orders",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    totalPages := (int(total) + perPage - 1) / perPage

    c.JSON(http.StatusOK, models.PaginatedResponse{
        Success: true,
        Message: "Orders retrieved successfully",
        Data:    orders,
        Meta: models.MetaData{
            CurrentPage: page,
            PerPage:     perPage,
            Total:       int(total),
            TotalPages:  totalPages,
        },
    })
}

// GetOrder retrieves a specific order by ID
func (h *OrderHandler) GetOrder(c *gin.Context) {
    orderID := c.Param("id")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("orders")
    var order bson.M
    err := collection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)

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

// CreateOrder creates a new order
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    userID, _, _, ok := middleware.GetUserFromContext(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, models.APIResponse{
            Success: false,
            Message: "Authentication required",
            Error:   stringPtr("auth_required"),
        })
        return
    }

    var req models.CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid request body",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    if len(req.Items) == 0 {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Order must contain at least one item",
            Error:   stringPtr("empty_order"),
        })
        return
    }

    orderID := uuid.New().String()
    orderNumber := fmt.Sprintf("ORD%s%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)

    var subtotal float64
    productsCollection := h.db.GetCollection("products")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var orderItems []bson.M
    for _, item := range req.Items {
        var product bson.M
        err := productsCollection.FindOne(ctx, bson.M{"_id": item.ProductID.String()}).Decode(&product)
        var price float64 = 0
        if err == nil {
            if p, ok := product["price"].(float64); ok {
                price = p
            }
        }
        totalPrice := price * float64(item.Quantity)
        subtotal += totalPrice

        orderItems = append(orderItems, bson.M{
            "product_id":            item.ProductID.String(),
            "quantity":              item.Quantity,
            "unit_price":            price,
            "total_price":           totalPrice,
            "special_instructions":  item.SpecialInstructions,
            "status":                "pending",
            "created_at":            time.Now(),
            "updated_at":            time.Now(),
        })
    }

    taxRate := 0.10
    taxAmount := subtotal * taxRate
    totalAmount := subtotal + taxAmount

    order := bson.M{
        "order_id":        orderID,
        "order_number":    orderNumber,
        "order_type":      req.OrderType,
        "status":          "pending",
        "subtotal":        subtotal,
        "tax_amount":      taxAmount,
        "discount_amount": 0,
        "total_amount":    totalAmount,
        "notes":           req.Notes,
        "user_id":         userID.String(),
        "created_at":      time.Now(),
        "updated_at":      time.Now(),
    }

    if req.TableID != nil {
        order["table_id"] = req.TableID.String()
    }
    if req.CustomerName != nil {
        order["customer_name"] = *req.CustomerName
    }

    ordersCollection := h.db.GetCollection("orders")
    orderResult, err := ordersCollection.InsertOne(ctx, order)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to create order",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    itemsCollection := h.db.GetCollection("order_items")
    for _, item := range orderItems {
        item["order_id"] = orderResult.InsertedID
        itemsCollection.InsertOne(ctx, item)
    }

    if req.OrderType == "dine_in" && req.TableID != nil {
        tablesCollection := h.db.GetCollection("tables")
        tablesCollection.UpdateOne(ctx,
            bson.M{"_id": req.TableID.String()},
            bson.M{"$set": bson.M{"is_occupied": true, "updated_at": time.Now()}},
        )
    }

    order["_id"] = orderResult.InsertedID

    c.JSON(http.StatusCreated, models.APIResponse{
        Success: true,
        Message: "Order created successfully",
        Data:    order,
    })
}

// UpdateOrderStatus updates the status of an order
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
    orderID := c.Param("id")

    var req models.UpdateOrderStatusRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid request body",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("orders")
    update := bson.M{
        "$set": bson.M{
            "status":     req.Status,
            "updated_at": time.Now(),
        },
    }

    if req.Status == "served" {
        update["$set"].(bson.M)["served_at"] = time.Now()
    } else if req.Status == "completed" {
        update["$set"].(bson.M)["completed_at"] = time.Now()
    }

    result, err := collection.UpdateOne(ctx, bson.M{"_id": orderID}, update)
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

    if req.Status == "completed" || req.Status == "cancelled" {
        tablesCollection := h.db.GetCollection("tables")
        var order bson.M
        collection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
        if tableID, ok := order["table_id"].(string); ok && tableID != "" {
            tablesCollection.UpdateOne(ctx,
                bson.M{"_id": tableID},
                bson.M{"$set": bson.M{"is_occupied": false, "updated_at": time.Now()}},
            )
        }
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Order status updated successfully",
    })
}
