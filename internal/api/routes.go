package api

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "strings"
    "time"

    "pos-backend/internal/database"
    "pos-backend/internal/handlers"
    "pos-backend/internal/middleware"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.RouterGroup, db *database.MongoDB, authMiddleware gin.HandlerFunc) {
    // Initialize handlers
    authHandler := handlers.NewAuthHandler(db)
    orderHandler := handlers.NewOrderHandler(db)
    productHandler := handlers.NewProductHandler(db)
    paymentHandler := handlers.NewPaymentHandler(db)
    tableHandler := handlers.NewTableHandler(db)

    // Public routes (no authentication required)
    public := router.Group("/")
    {
        public.POST("/auth/login", authHandler.Login)
        public.POST("/auth/logout", authHandler.Logout)
    }

    // Protected routes (authentication required)
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
        protected.GET("/orders", orderHandler.GetOrders)
        protected.GET("/orders/:id", orderHandler.GetOrder)
        protected.PATCH("/orders/:id/status", orderHandler.UpdateOrderStatus)
        protected.GET("/orders/:id/payments", paymentHandler.GetPayments)
        protected.GET("/orders/:id/payment-summary", paymentHandler.GetPaymentSummary)
    }

    // Server routes (server role - dine-in orders only)
    server := router.Group("/server")
    server.Use(authMiddleware)
    server.Use(middleware.RequireRole("server"))
    {
        server.POST("/orders", createDineInOrder(db))
    }

    // Counter routes (counter role - all order types and payments)
    counter := router.Group("/counter")
    counter.Use(authMiddleware)
    counter.Use(middleware.RequireRole("counter"))
    {
        counter.POST("/orders", orderHandler.CreateOrder)
        counter.POST("/orders/:id/payments", paymentHandler.ProcessPayment)
    }

    // Admin routes (admin/manager only)
    admin := router.Group("/admin")
    admin.Use(authMiddleware)
    admin.Use(middleware.RequireRoles([]string{"admin", "manager"}))
    {
        admin.GET("/dashboard/stats", getDashboardStats(db))
        admin.GET("/reports/sales", getSalesReport(db))
        admin.GET("/reports/orders", getOrdersReport(db))
        admin.GET("/reports/income", getIncomeReport(db))
    }
}

// Dashboard stats handler
func getDashboardStats(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        stats := make(map[string]interface{})

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        ordersCollection := db.GetCollection("orders")
        
        // Today's orders count
        startOfDay := time.Now().UTC().Truncate(24 * time.Hour)
        todayFilter := bson.M{"created_at": bson.M{"$gte": startOfDay}}
        todayOrders, _ := ordersCollection.CountDocuments(ctx, todayFilter)
        
        // Today's revenue
        matchStage := bson.D{{Key: "$match", Value: bson.M{
            "status": "completed",
            "created_at": bson.M{"$gte": startOfDay},
        }}}
        groupStage := bson.D{{Key: "$group", Value: bson.M{
            "_id": nil,
            "total": bson.M{"$sum": "$total_amount"},
        }}}
        
        cursor, err := ordersCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage})
        var todayRevenue float64 = 0
        if err == nil {
            var results []bson.M
            if err = cursor.All(ctx, &results); err == nil && len(results) > 0 {
                if total, ok := results[0]["total"].(float64); ok {
                    todayRevenue = total
                }
            }
        }
        
        // Active orders (not completed or cancelled)
        activeFilter := bson.M{"status": bson.M{"$nin": []string{"completed", "cancelled"}}}
        activeOrders, _ := ordersCollection.CountDocuments(ctx, activeFilter)
        
        // Occupied tables
        tablesCollection := db.GetCollection("tables")
        occupiedTables, _ := tablesCollection.CountDocuments(ctx, bson.M{"is_occupied": true})

        stats["today_orders"] = todayOrders
        stats["today_revenue"] = todayRevenue
        stats["active_orders"] = activeOrders
        stats["occupied_tables"] = occupiedTables

        c.JSON(200, gin.H{
            "success": true,
            "message": "Dashboard stats retrieved successfully",
            "data":    stats,
        })
    }
}

// Sales report handler
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
        
        collection := db.GetCollection("orders")
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        
        matchStage := bson.D{{Key: "$match", Value: bson.M{
            "status": "completed",
            "created_at": bson.M{"$gte": startDate},
        }}}
        
        groupStage := bson.D{{Key: "$group", Value: bson.M{
            "_id": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"}},
            "order_count": bson.M{"$sum": 1},
            "revenue": bson.M{"$sum": "$total_amount"},
        }}}
        
        sortStage := bson.D{{Key: "$sort", Value: bson.M{"_id": -1}}}
        
        cursor, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, sortStage})
        if err != nil {
            c.JSON(500, gin.H{
                "success": false,
                "message": "Failed to fetch sales report",
                "error":   err.Error(),
            })
            return
        }
        defer cursor.Close(ctx)
        
        var report []map[string]interface{}
        for cursor.Next(ctx) {
            var result bson.M
            if err := cursor.Decode(&result); err != nil {
                continue
            }
            report = append(report, map[string]interface{}{
                "date":        result["_id"],
                "order_count": result["order_count"],
                "revenue":     result["revenue"],
            })
        }
        
        c.JSON(200, gin.H{
            "success": true,
            "message": "Sales report retrieved successfully",
            "data":    report,
        })
    }
}

// Orders report handler
func getOrdersReport(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        collection := db.GetCollection("orders")
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        
        matchStage := bson.D{{Key: "$match", Value: bson.M{
            "created_at": bson.M{"$gte": time.Now().UTC().Truncate(24 * time.Hour)},
        }}}
        
        groupStage := bson.D{{Key: "$group", Value: bson.M{
            "_id": "$status",
            "count": bson.M{"$sum": 1},
            "avg_amount": bson.M{"$avg": "$total_amount"},
        }}}
        
        cursor, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage})
        if err != nil {
            c.JSON(500, gin.H{
                "success": false,
                "message": "Failed to fetch orders report",
                "error":   err.Error(),
            })
            return
        }
        defer cursor.Close(ctx)
        
        var report []map[string]interface{}
        for cursor.Next(ctx) {
            var result bson.M
            if err := cursor.Decode(&result); err != nil {
                continue
            }
            report = append(report, map[string]interface{}{
                "status":     result["_id"],
                "count":      result["count"],
                "avg_amount": result["avg_amount"],
            })
        }
        
        c.JSON(200, gin.H{
            "success": true,
            "message": "Orders report retrieved successfully",
            "data":    report,
        })
    }
}

// Income report handler
func getIncomeReport(db *database.MongoDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        period := c.DefaultQuery("period", "today")
        
        var startDate time.Time
        now := time.Now().UTC()
        var dateFormat string
        
        switch period {
        case "week":
            startDate = now.AddDate(0, 0, -7)
            dateFormat = "%Y-%m-%d"
        case "month":
            startDate = now.AddDate(0, -1, 0)
            dateFormat = "%Y-%m-%d"
        case "year":
            startDate = now.AddDate(-1, 0, 0)
            dateFormat = "%Y-%m"
        default:
            startDate = now.Truncate(24 * time.Hour)
            dateFormat = "%Y-%m-%d %H:00"
        }
        
        collection := db.GetCollection("orders")
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        
        matchStage := bson.D{{Key: "$match", Value: bson.M{
            "status": "completed",
            "created_at": bson.M{"$gte": startDate},
        }}}
        
        groupStage := bson.D{{Key: "$group", Value: bson.M{
            "_id": bson.M{"$dateToString": bson.M{"format": dateFormat, "date": "$created_at"}},
            "total_orders": bson.M{"$sum": 1},
            "gross_income": bson.M{"$sum": "$total_amount"},
            "tax_collected": bson.M{"$sum": "$tax_amount"},
            "net_income": bson.M{"$sum": bson.M{"$subtract": []interface{}{"$total_amount", "$tax_amount"}}},
        }}}
        
        sortStage := bson.D{{Key: "$sort", Value: bson.M{"_id": -1}}}
        
        cursor, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, sortStage})
        if err != nil {
            c.JSON(500, gin.H{
                "success": false,
                "message": "Failed to fetch income report",
                "error":   err.Error(),
            })
            return
        }
        defer cursor.Close(ctx)
        
        var breakdown []map[string]interface{}
        var totalOrders int
        var totalGross, totalTax, totalNet float64
        
        for cursor.Next(ctx) {
            var result bson.M
            if err := cursor.Decode(&result); err != nil {
                continue
            }
            
            periodStr := fmt.Sprintf("%v", result["_id"])
            orders := toInt(result["total_orders"])
            gross := toFloat64(result["gross_income"])
            tax := toFloat64(result["tax_collected"])
            net := toFloat64(result["net_income"])
            
            totalOrders += orders
            totalGross += gross
            totalTax += tax
            totalNet += net
            
            breakdown = append(breakdown, map[string]interface{}{
                "period": periodStr,
                "orders": orders,
                "gross":  gross,
                "tax":    tax,
                "net":    net,
            })
        }
        
        result := map[string]interface{}{
            "summary": map[string]interface{}{
                "total_orders":  totalOrders,
                "gross_income":  totalGross,
                "tax_collected": totalTax,
                "net_income":    totalNet,
            },
            "breakdown": breakdown,
            "period":    period,
        }
        
        c.JSON(200, gin.H{
            "success": true,
            "message": "Income report retrieved successfully",
            "data":    result,
        })
    }
}

// Server role handler - only allows dine-in orders
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
            c.JSON(400, gin.H{
                "success": false,
                "message": "Invalid request body",
                "error":   err.Error(),
            })
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

// Helper functions
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
