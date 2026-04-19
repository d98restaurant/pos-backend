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

// GetTables retrieves all dining tables
func (h *TableHandler) GetTables(c *gin.Context) {
    location := c.Query("location")
    occupiedOnly := c.Query("occupied_only") == "true"
    availableOnly := c.Query("available_only") == "true"

    filter := bson.M{}
    if location != "" {
        filter["location"] = bson.M{"$regex": location, "$options": "i"}
    }
    if occupiedOnly {
        filter["is_occupied"] = true
    } else if availableOnly {
        filter["is_occupied"] = false
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("tables")
    cursor, err := collection.Find(ctx, filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to fetch tables",
            Error:   stringPtr(err.Error()),
        })
        return
    }
    defer cursor.Close(ctx)

    var tables []bson.M
    if err = cursor.All(ctx, &tables); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to scan table",
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

// GetTable retrieves a specific table by ID
func (h *TableHandler) GetTable(c *gin.Context) {
    tableID := c.Param("id")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("tables")
    var table bson.M
    err := collection.FindOne(ctx, bson.M{"_id": tableID}).Decode(&table)

    if err != nil {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Table not found",
            Error:   stringPtr("table_not_found"),
        })
        return
    }

    ordersCollection := h.db.GetCollection("orders")
    var order bson.M
    ordersCollection.FindOne(ctx, bson.M{"table_id": tableID, "status": bson.M{"$nin": []string{"completed", "cancelled"}}}).Decode(&order)

    response := map[string]interface{}{
        "id":               table["_id"],
        "table_number":     table["table_number"],
        "seating_capacity": table["seating_capacity"],
        "location":         table["location"],
        "is_occupied":      table["is_occupied"],
        "created_at":       table["created_at"],
        "updated_at":       table["updated_at"],
        "current_order":    order,
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Table retrieved successfully",
        Data:    response,
    })
}

// GetTablesByLocation retrieves tables grouped by location
func (h *TableHandler) GetTablesByLocation(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("tables")
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

    var tables []bson.M
    cursor.All(ctx, &tables)

    locationMap := make(map[string][]bson.M)
    for _, table := range tables {
        location := "General"
        if loc, ok := table["location"].(string); ok && loc != "" {
            location = loc
        }
        locationMap[location] = append(locationMap[location], table)
    }

    var locations []map[string]interface{}
    for locationName, locationTables := range locationMap {
        locations = append(locations, map[string]interface{}{
            "location": locationName,
            "tables":   locationTables,
        })
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Tables grouped by location retrieved successfully",
        Data:    locations,
    })
}

// GetTableStatus retrieves the status overview of all tables
func (h *TableHandler) GetTableStatus(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("tables")
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to fetch table status",
            Error:   stringPtr(err.Error()),
        })
        return
    }
    defer cursor.Close(ctx)

    var tables []bson.M
    cursor.All(ctx, &tables)

    totalTables := len(tables)
    occupiedTables := 0
    locationStatsMap := make(map[string]map[string]int)

    for _, table := range tables {
        isOccupied := false
        if occupied, ok := table["is_occupied"].(bool); ok && occupied {
            isOccupied = true
            occupiedTables++
        }

        location := "General"
        if loc, ok := table["location"].(string); ok && loc != "" {
            location = loc
        }

        if _, exists := locationStatsMap[location]; !exists {
            locationStatsMap[location] = map[string]int{"total": 0, "occupied": 0}
        }
        locationStatsMap[location]["total"]++
        if isOccupied {
            locationStatsMap[location]["occupied"]++
        }
    }

    var locationStats []map[string]interface{}
    for location, stats := range locationStatsMap {
        total := stats["total"]
        occupied := stats["occupied"]
        locationStats = append(locationStats, map[string]interface{}{
            "location":         location,
            "total_tables":     total,
            "occupied_tables":  occupied,
            "available_tables": total - occupied,
            "occupancy_rate":   float64(occupied) / float64(total) * 100,
        })
    }

    response := map[string]interface{}{
        "total_tables":     totalTables,
        "occupied_tables":  occupiedTables,
        "available_tables": totalTables - occupiedTables,
        "occupancy_rate":   float64(occupiedTables) / float64(totalTables) * 100,
        "by_location":      locationStats,
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Table status retrieved successfully",
        Data:    response,
    })
}

// GetAdminTables - Admin endpoint for tables
func (h *TableHandler) GetAdminTables(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("tables")
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

    var tables []bson.M
    cursor.All(ctx, &tables)

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Tables retrieved successfully",
        Data:    tables,
    })
}

// CreateTable creates a new table
func (h *TableHandler) CreateTable(c *gin.Context) {
    var req struct {
        TableNumber     string  `json:"table_number" binding:"required"`
        SeatingCapacity int     `json:"seating_capacity"`
        Location        *string `json:"location"`
    }

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

    collection := h.db.GetCollection("tables")
    result, err := collection.InsertOne(ctx, bson.M{
        "table_number":     req.TableNumber,
        "seating_capacity": req.SeatingCapacity,
        "location":         req.Location,
        "is_occupied":      false,
        "created_at":       time.Now(),
        "updated_at":       time.Now(),
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to create table",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    c.JSON(http.StatusCreated, models.APIResponse{
        Success: true,
        Message: "Table created successfully",
        Data:    map[string]interface{}{"id": result.InsertedID},
    })
}

// UpdateTable updates an existing table
func (h *TableHandler) UpdateTable(c *gin.Context) {
    tableID := c.Param("id")

    var req struct {
        TableNumber     *string `json:"table_number"`
        SeatingCapacity *int    `json:"seating_capacity"`
        Location        *string `json:"location"`
        IsOccupied      *bool   `json:"is_occupied"`
    }

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

    update := bson.M{"$set": bson.M{"updated_at": time.Now()}}
    if req.TableNumber != nil {
        update["$set"].(bson.M)["table_number"] = *req.TableNumber
    }
    if req.SeatingCapacity != nil {
        update["$set"].(bson.M)["seating_capacity"] = *req.SeatingCapacity
    }
    if req.Location != nil {
        update["$set"].(bson.M)["location"] = *req.Location
    }
    if req.IsOccupied != nil {
        update["$set"].(bson.M)["is_occupied"] = *req.IsOccupied
    }

    collection := h.db.GetCollection("tables")
    result, err := collection.UpdateOne(ctx, bson.M{"_id": tableID}, update)

    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to update table",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Table not found",
            Error:   stringPtr("table_not_found"),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Table updated successfully",
    })
}

// DeleteTable deletes a table
func (h *TableHandler) DeleteTable(c *gin.Context) {
    tableID := c.Param("id")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("tables")
    result, err := collection.DeleteOne(ctx, bson.M{"_id": tableID})

    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to delete table",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    if result.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Table not found",
            Error:   stringPtr("table_not_found"),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Table deleted successfully",
    })
}
