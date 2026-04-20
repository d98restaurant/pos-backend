package api

import (
    "context"
    "encoding/json"
    "io"
    "strings"
    "time"

    "pos-backend/internal/database"
    "pos-backend/internal/handlers"
    "pos-backend/internal/middleware"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.RouterGroup, db *database.MongoDB, authMiddleware gin.HandlerFunc) {
    authHandler := handlers.NewAuthHandler(db)
    orderHandler := handlers.NewOrderHandler(db)
    productHandler := handlers.NewProductHandler(db)
    paymentHandler := handlers.NewPaymentHandler(db)
    tableHandler := handlers.NewTableHandler(db)

    public := router.Group("/")
    {
        public.POST("/auth/login", authHandler.Login)
        public.POST("/auth/logout", authHandler.Logout)
    }

    protected := router.Group("/")
    protected.Use(authMiddleware)
    {
        protected.GET("/auth/me", authHandler.GetCurrentUser)
        protected.GET("/products", productHandler.GetProducts)
        protected.GET("/products/:id", productHandler.GetProduct)
        protected.GET("/categories", productHandler.GetCategories)
        protected.GET("/categories/:id/products", productHandler.GetProductsByCategory)
        protected.GET("/tables", tableHandler.GetTables)
        protected.GET("/tables/:id", tableHandler.GetTable)
        protected.GET("/tables/by-location", tableHandler.GetTablesByLocation)
        protected.GET("/tables/status", tableHandler.GetTableStatus)
        protected.GET("/orders", orderHandler.GetOrders)
        protected.GET("/orders/:id", orderHandler.GetOrder)
        protected.PATCH("/orders/:id/status", orderHandler.UpdateOrderStatus)
        protected.GET("/orders/:id/payments", paymentHandler.GetPayments)
        protected.GET("/orders/:id/payment-summary", paymentHandler.GetPaymentSummary)
    }

    // Server routes - Allow server, admin, and manager roles
    server := router.Group("/server")
    server.Use(authMiddleware)
    server.Use(middleware.RequireRoles([]string{"server", "admin", "manager"}))
    {
        server.POST("/orders", createDineInOrder(db))
    }

    // Counter routes - Allow counter, admin, and manager roles
    counter := router.Group("/counter")
    counter.Use(authMiddleware)
    counter.Use(middleware.RequireRoles([]string{"counter", "admin", "manager"}))
    {
        counter.POST("/orders", orderHandler.CreateOrder)
        counter.POST("/orders/:id/payments", paymentHandler.ProcessPayment)
    }

    // Admin routes - Only admin and manager
    admin := router.Group("/admin")
    admin.Use(authMiddleware)
    admin.Use(middleware.RequireRoles([]string{"admin", "manager"}))
    {
        admin.GET("/dashboard/stats", getDashboardStats(db))
        admin.GET("/reports/sales", getSalesReport(db))
        admin.GET("/reports/orders", getOrdersReport(db))
        admin.GET("/reports/income", getIncomeReport(db))
        admin.GET("/products", productHandler.GetAdminProducts)
        admin.GET("/categories", productHandler.GetAdminCategories)
        admin.POST("/categories", productHandler.CreateCategory)
        admin.PUT("/categories/:id", productHandler.UpdateCategory)
        admin.DELETE("/categories/:id", productHandler.DeleteCategory)
        admin.POST("/products", productHandler.CreateProduct)
        admin.PUT("/products/:id", productHandler.UpdateProduct)
        admin.DELETE("/products/:id", productHandler.DeleteProduct)
        admin.GET("/tables", tableHandler.GetAdminTables)
        admin.POST("/tables", tableHandler.CreateTable)
        admin.PUT("/tables/:id", tableHandler.UpdateTable)
        admin.DELETE("/tables/:id", tableHandler.DeleteTable)
        admin.GET("/users", getAdminUsers(db))
        admin.POST("/users", createUser(db))
        admin.PUT("/users/:id", updateUser(db))
        admin.DELETE("/users/:id", deleteUser(db))
        admin.POST("/orders", orderHandler.CreateOrder)
        admin.POST("/orders/:id/payments", paymentHandler.ProcessPayment)
    }

    // Kitchen routes - Allow kitchen, admin, and manager roles
    kitchen := router.Group("/kitchen")
    kitchen.Use(authMiddleware)
    kitchen.Use(middleware.RequireRoles([]string{"kitchen", "admin", "manager"}))
    {
        kitchen.GET("/orders", getKitchenOrders(db))
        kitchen.PATCH("/orders/:id/items/:item_id/status", updateOrderItemStatus(db))
    }
}

func getDashboardStats(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        ordersCollection := db.GetCollection("orders")
        tablesCollection := db.GetCollection("tables")

        startOfDay := time.Now().UTC().Truncate(24 * time.Hour)
        todayOrders, _ := ordersCollection.CountDocuments(ctx, bson.M{"created_at": bson.M{"$gte": startOfDay}})

        cursor, _ := ordersCollection.Aggregate(ctx, bson.A{
            bson.M{"$match": bson.M{"status": "completed", "created_at": bson.M{"$gte": startOfDay}}},
            bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$total_amount"}}},
        })
        var todayRevenue float64 = 0
        var result []bson.M
        if err := cursor.All(ctx, &result); err == nil && len(result) > 0 {
            if total, ok := result[0]["total"].(float64); ok {
                todayRevenue = total
            }
        }

        activeOrders, _ := ordersCollection.CountDocuments(ctx, bson.M{"status": bson.M{"$nin": []string{"completed", "cancelled"}}})
        occupiedTables, _ := tablesCollection.CountDocuments(ctx, bson.M{"is_occupied": true})

        c.JSON(200, gin.H{
            "success": true,
            "message": "Dashboard stats retrieved successfully",
            "data": gin.H{
                "today_orders":    todayOrders,
                "today_revenue":   todayRevenue,
                "active_orders":   activeOrders,
                "occupied_tables": occupiedTables,
            },
        })
    }
}

func getSalesReport(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        period := c.DefaultQuery("period", "today")
        var startDate time.Time
        now := time.Now().UTC()

        switch period {
        case "week":
            startDate = now.AddDate(0, 0, -7)
        case "month":
            startDate = now.AddDate(0, -1, 0)
        default:
            startDate = now.Truncate(24 * time.Hour)
        }

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        collection := db.GetCollection("orders")
        cursor, err := collection.Aggregate(ctx, bson.A{
            bson.M{"$match": bson.M{"status": "completed", "created_at": bson.M{"$gte": startDate}}},
            bson.M{"$group": bson.M{
                "_id":         bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"}},
                "order_count": bson.M{"$sum": 1},
                "revenue":     bson.M{"$sum": "$total_amount"},
            }},
            bson.M{"$sort": bson.M{"_id": -1}},
        })
        if err != nil {
            c.JSON(500, gin.H{"success": false, "message": "Failed to fetch sales report", "error": err.Error()})
            return
        }
        defer cursor.Close(ctx)

        var report []gin.H
        for cursor.Next(ctx) {
            var result bson.M
            cursor.Decode(&result)
            report = append(report, gin.H{
                "date":        result["_id"],
                "order_count": result["order_count"],
                "revenue":     result["revenue"],
            })
        }

        c.JSON(200, gin.H{"success": true, "message": "Sales report retrieved successfully", "data": report})
    }
}

func getOrdersReport(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        collection := db.GetCollection("orders")
        cursor, err := collection.Aggregate(ctx, bson.A{
            bson.M{"$match": bson.M{"created_at": bson.M{"$gte": time.Now().UTC().Truncate(24 * time.Hour)}}},
            bson.M{"$group": bson.M{
                "_id":        "$status",
                "count":      bson.M{"$sum": 1},
                "avg_amount": bson.M{"$avg": "$total_amount"},
            }},
        })
        if err != nil {
            c.JSON(500, gin.H{"success": false, "message": "Failed to fetch orders report", "error": err.Error()})
            return
        }
        defer cursor.Close(ctx)

        var report []gin.H
        for cursor.Next(ctx) {
            var result bson.M
            cursor.Decode(&result)
            report = append(report, gin.H{
                "status":     result["_id"],
                "count":      result["count"],
                "avg_amount": result["avg_amount"],
            })
        }

        c.JSON(200, gin.H{"success": true, "message": "Orders report retrieved successfully", "data": report})
    }
}

func getIncomeReport(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        period := c.DefaultQuery("period", "today")
        var startDate time.Time
        now := time.Now().UTC()

        switch period {
        case "week":
            startDate = now.AddDate(0, 0, -7)
        case "month":
            startDate = now.AddDate(0, -1, 0)
        case "year":
            startDate = now.AddDate(-1, 0, 0)
        default:
            startDate = now.Truncate(24 * time.Hour)
        }

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        collection := db.GetCollection("orders")
        cursor, err := collection.Aggregate(ctx, bson.A{
            bson.M{"$match": bson.M{"status": "completed", "created_at": bson.M{"$gte": startDate}}},
            bson.M{"$group": bson.M{
                "_id":            bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"}},
                "total_orders":   bson.M{"$sum": 1},
                "gross_income":   bson.M{"$sum": "$total_amount"},
                "tax_collected":  bson.M{"$sum": "$tax_amount"},
                "net_income":     bson.M{"$sum": bson.M{"$subtract": []interface{}{"$total_amount", "$tax_amount"}}},
            }},
            bson.M{"$sort": bson.M{"_id": -1}},
        })
        if err != nil {
            c.JSON(500, gin.H{"success": false, "message": "Failed to fetch income report", "error": err.Error()})
            return
        }
        defer cursor.Close(ctx)

        var breakdown []gin.H
        var totalOrders int
        var totalGross, totalTax, totalNet float64

        for cursor.Next(ctx) {
            var result bson.M
            cursor.Decode(&result)
            orders := toInt(result["total_orders"])
            gross := toFloat64(result["gross_income"])
            tax := toFloat64(result["tax_collected"])
            net := toFloat64(result["net_income"])

            totalOrders += orders
            totalGross += gross
            totalTax += tax
            totalNet += net

            breakdown = append(breakdown, gin.H{
                "period": result["_id"],
                "orders": orders,
                "gross":  gross,
                "tax":    tax,
                "net":    net,
            })
        }

        c.JSON(200, gin.H{
            "success": true,
            "message": "Income report retrieved successfully",
            "data": gin.H{
                "summary": gin.H{
                    "total_orders":  totalOrders,
                    "gross_income":  totalGross,
                    "tax_collected": totalTax,
                    "net_income":    totalNet,
                },
                "breakdown": breakdown,
                "period":    period,
            },
        })
    }
}

func getKitchenOrders(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        status := c.DefaultQuery("status", "all")

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        filter := bson.M{"status": bson.M{"$in": []string{"confirmed", "preparing", "ready"}}}
        if status != "all" {
            filter["status"] = status
        }

        collection := db.GetCollection("orders")
        cursor, err := collection.Find(ctx, filter)
        if err != nil {
            c.JSON(500, gin.H{"success": false, "message": "Failed to fetch kitchen orders", "error": err.Error()})
            return
        }
        defer cursor.Close(ctx)

        var orders []gin.H
        for cursor.Next(ctx) {
            var order bson.M
            cursor.Decode(&order)
            orders = append(orders, gin.H{
                "id":            order["_id"],
                "order_number":  order["order_number"],
                "table_id":      order["table_id"],
                "order_type":    order["order_type"],
                "status":        order["status"],
                "customer_name": order["customer_name"],
                "created_at":    order["created_at"],
            })
        }

        c.JSON(200, gin.H{"success": true, "message": "Kitchen orders retrieved successfully", "data": orders})
    }
}

func updateOrderItemStatus(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        orderID := c.Param("id")
        itemID := c.Param("item_id")

        var req struct {
            Status string `json:"status"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"success": false, "message": "Invalid request body", "error": err.Error()})
            return
        }

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        collection := db.GetCollection("order_items")
        _, err := collection.UpdateOne(ctx,
            bson.M{"_id": itemID, "order_id": orderID},
            bson.M{"$set": bson.M{"status": req.Status, "updated_at": time.Now()}},
        )

        if err != nil {
            c.JSON(500, gin.H{"success": false, "message": "Failed to update order item status", "error": err.Error()})
            return
        }

        c.JSON(200, gin.H{"success": true, "message": "Order item status updated successfully"})
    }
}

func createDineInOrder(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req struct {
            TableID      *string `json:"table_id"`
            CustomerName *string `json:"customer_name"`
            Items        []struct {
                ProductID           string  `json:"product_id"`
                Quantity            int     `json:"quantity"`
                SpecialInstructions *string `json:"special_instructions"`
            } `json:"items"`
            Notes *string `json:"notes"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"success": false, "message": "Invalid request body", "error": err.Error()})
            return
        }

        orderHandler := handlers.NewOrderHandler(db)

        createOrderReq := map[string]interface{}{
            "table_id":      req.TableID,
            "customer_name": req.CustomerName,
            "order_type":    "dine_in",
            "items":         req.Items,
            "notes":         req.Notes,
        }

        reqBytes, _ := json.Marshal(createOrderReq)
        c.Request.Body = io.NopCloser(strings.NewReader(string(reqBytes)))

        orderHandler.CreateOrder(c)
    }
}

func getAdminUsers(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        collection := db.GetCollection("users")
        cursor, err := collection.Find(ctx, bson.M{})
        if err != nil {
            c.JSON(500, gin.H{"success": false, "message": "Failed to fetch users", "error": err.Error()})
            return
        }
        defer cursor.Close(ctx)

        var users []bson.M
        cursor.All(ctx, &users)

        c.JSON(200, gin.H{"success": true, "message": "Users retrieved successfully", "data": users})
    }
}

func createUser(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req struct {
            Username  string `json:"username" binding:"required"`
            Email     string `json:"email" binding:"required"`
            Password  string `json:"password" binding:"required"`
            FirstName string `json:"first_name" binding:"required"`
            LastName  string `json:"last_name" binding:"required"`
            Role      string `json:"role" binding:"required"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"success": false, "message": "Invalid request body", "error": err.Error()})
            return
        }

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        collection := db.GetCollection("users")
        result, err := collection.InsertOne(ctx, bson.M{
            "username":    req.Username,
            "email":       req.Email,
            "first_name":  req.FirstName,
            "last_name":   req.LastName,
            "role":        req.Role,
            "is_active":   true,
            "created_at":  time.Now(),
            "updated_at":  time.Now(),
        })

        if err != nil {
            c.JSON(500, gin.H{"success": false, "message": "Failed to create user", "error": err.Error()})
            return
        }

        c.JSON(201, gin.H{"success": true, "message": "User created successfully", "data": gin.H{"id": result.InsertedID}})
    }
}

func updateUser(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.Param("id")
        var req map[string]interface{}
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"success": false, "message": "Invalid request body", "error": err.Error()})
            return
        }

        req["updated_at"] = time.Now()

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        collection := db.GetCollection("users")
        result, err := collection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": req})
        if err != nil {
            c.JSON(500, gin.H{"success": false, "message": "Failed to update user", "error": err.Error()})
            return
        }
        if result.MatchedCount == 0 {
            c.JSON(404, gin.H{"success": false, "message": "User not found"})
            return
        }

        c.JSON(200, gin.H{"success": true, "message": "User updated successfully"})
    }
}

func deleteUser(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.Param("id")

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        collection := db.GetCollection("users")
        result, err := collection.DeleteOne(ctx, bson.M{"_id": userID})
        if err != nil {
            c.JSON(500, gin.H{"success": false, "message": "Failed to delete user", "error": err.Error()})
            return
        }
        if result.DeletedCount == 0 {
            c.JSON(404, gin.H{"success": false, "message": "User not found"})
            return
        }

        c.JSON(200, gin.H{"success": true, "message": "User deleted successfully"})
    }
}

func toInt(v interface{}) int {
    switch val := v.(type) {
    case int:
        return val
    case int32:
        return int(val)
    case int64:
        return int(val)
    case float64:
        return int(val)
    default:
        return 0
    }
}

func toFloat64(v interface{}) float64 {
    switch val := v.(type) {
    case float64:
        return val
    case int:
        return float64(val)
    case int32:
        return float64(val)
    case int64:
        return float64(val)
    default:
        return 0
    }
}