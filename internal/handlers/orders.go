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
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
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
        // Handle comma-separated statuses
        statuses := splitStatuses(status)
        if len(statuses) > 1 {
            filter["status"] = bson.M{"$in": statuses}
        } else {
            filter["status"] = status
        }
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

    // Fetch order items for each order
    itemsCollection := h.db.GetCollection("order_items")
    for i, order := range orders {
        orderID := order["_id"].(primitive.ObjectID).Hex()
        cursor, err := itemsCollection.Find(ctx, bson.M{"order_id": orderID})
        if err == nil {
            var items []bson.M
            cursor.All(ctx, &items)
            orders[i]["items"] = items
        }
        cursor.Close(ctx)
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

// Helper function to split comma-separated statuses
func splitStatuses(status string) []string {
    var result []string
    for _, s := range splitString(status, ",") {
        if s != "" {
            result = append(result, s)
        }
    }
    return result
}

func splitString(s, sep string) []string {
    var result []string
    start := 0
    for i := 0; i < len(s); i++ {
        if s[i] == sep[0] {
            result = append(result, s[start:i])
            start = i + 1
        }
    }
    result = append(result, s[start:])
    return result
}

// GetOrder retrieves a specific order by ID with items
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

    // Fetch order items
    itemsCollection := h.db.GetCollection("order_items")
    cursor, err := itemsCollection.Find(ctx, bson.M{"order_id": orderID})
    if err == nil {
        var items []bson.M
        cursor.All(ctx, &items)
        order["items"] = items
    }
    cursor.Close(ctx)

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

    // Generate ObjectID for order
    orderID := primitive.NewObjectID()
    orderNumber := fmt.Sprintf("ORD%s%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)

    var subtotal float64 = 0
    productsCollection := h.db.GetCollection("products")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var orderItems []bson.M
    // In the CreateOrder function, update the orderItems creation:
for _, item := range req.Items {
    var product bson.M
    err := productsCollection.FindOne(ctx, bson.M{"_id": item.ProductID}).Decode(&product)
    
    var price float64 = 0
    var productName string = "Unknown Product"
    
    if err == nil {
        if p, ok := product["price"].(float64); ok {
            price = p
        }
        if n, ok := product["name"].(string); ok {
            productName = n
        }
    }
    
    totalPrice := price * float64(item.Quantity)
    subtotal += totalPrice

    orderItems = append(orderItems, bson.M{
        "product_id":            item.ProductID,
        "product_name":          productName,  // Make sure this is set
        "name":                  productName,  // Also store as name for easier access
        "quantity":              item.Quantity,
        "unit_price":            price,
        "total_price":           totalPrice,
        "special_instructions":  item.SpecialInstructions,
        "status":                "pending",
        "created_at":            time.Now(),
        "updated_at":            time.Now(),
    })
}

    // Calculate taxes (10% tax rate)
    taxRate := 0.10
    taxAmount := subtotal * taxRate
    totalAmount := subtotal + taxAmount

    // Set status to "confirmed" so kitchen can see it
    initialStatus := "confirmed"

    order := bson.M{
        "_id":             orderID,
        "order_number":    orderNumber,
        "order_type":      req.OrderType,
        "status":          initialStatus,
        "subtotal":        subtotal,
        "tax_amount":      taxAmount,
        "discount_amount": 0,
        "total_amount":    totalAmount,
        "notes":           req.Notes,
        "user_id":         userID,
        "created_at":      time.Now(),
        "updated_at":      time.Now(),
    }

    if req.TableID != nil {
        order["table_id"] = *req.TableID
    }
    if req.CustomerName != nil {
        order["customer_name"] = *req.CustomerName
    }

    ordersCollection := h.db.GetCollection("orders")
    _, err := ordersCollection.InsertOne(ctx, order)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to create order",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    // Insert order items
    itemsCollection := h.db.GetCollection("order_items")
    for _, item := range orderItems {
        item["order_id"] = orderID.Hex()
        itemsCollection.InsertOne(ctx, item)
    }

    // Update table occupancy for dine-in orders
    if req.OrderType == "dine_in" && req.TableID != nil {
        tablesCollection := h.db.GetCollection("tables")
        tablesCollection.UpdateOne(ctx,
            bson.M{"_id": *req.TableID},
            bson.M{"$set": bson.M{"is_occupied": true, "updated_at": time.Now()}},
        )
    }

    // Return the created order with items
    order["_id"] = orderID.Hex()
    order["items"] = orderItems

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

    // If order is completed or cancelled, free the table
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