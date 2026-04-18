package handlers

import (
    "context"
    "net/http"
    "strconv"
    "time"

    "pos-backend/internal/database"
    "pos-backend/internal/models"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type ProductHandler struct {
    db *database.MongoDB
}

func NewProductHandler(db *database.MongoDB) *ProductHandler {
    return &ProductHandler{db: db}
}

// GetProducts retrieves all products with pagination and filtering
func (h *ProductHandler) GetProducts(c *gin.Context) {
    // Parse query parameters
    page := 1
    perPage := 50
    categoryID := c.Query("category_id")
    available := c.Query("available")
    search := c.Query("search")

    if pageStr := c.Query("page"); pageStr != "" {
        if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
            page = p
        }
    }

    if perPageStr := c.Query("per_page"); perPageStr != "" {
        if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
            perPage = pp
        }
    }

    offset := int64((page - 1) * perPage)

    // Build filter
    filter := bson.M{}
    
    if categoryID != "" {
        filter["category_id"] = categoryID
    }
    
    if available == "true" {
        filter["is_available"] = true
    } else if available == "false" {
        filter["is_available"] = false
    }
    
    if search != "" {
        filter["$or"] = []bson.M{
            {"name": bson.M{"$regex": search, "$options": "i"}},
            {"description": bson.M{"$regex": search, "$options": "i"}},
        }
    }

    collection := h.db.GetCollection("products")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Count total records
    total, err := collection.CountDocuments(ctx, filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to count products",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    // Find with pagination
    findOptions := options.Find()
    findOptions.SetSort(bson.D{{Key: "sort_order", Value: 1}, {Key: "name", Value: 1}})
    findOptions.SetLimit(int64(perPage))
    findOptions.SetSkip(offset)

    cursor, err := collection.Find(ctx, filter, findOptions)
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

    totalPages := (int(total) + perPage - 1) / perPage

    c.JSON(http.StatusOK, models.PaginatedResponse{
        Success: true,
        Message: "Products retrieved successfully",
        Data:    products,
        Meta: models.MetaData{
            CurrentPage: page,
            PerPage:     perPage,
            Total:       int(total),
            TotalPages:  totalPages,
        },
    })
}

// GetProduct retrieves a specific product by ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
    productID := c.Param("id")
    
    collection := h.db.GetCollection("products")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var product models.MongoProduct
    err := collection.FindOne(ctx, bson.M{"product_id": productID}).Decode(&product)
    
    if err != nil {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Product not found",
            Error:   stringPtr("product_not_found"),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Product retrieved successfully",
        Data:    product,
    })
}

// GetCategories retrieves all categories
func (h *ProductHandler) GetCategories(c *gin.Context) {
    activeOnly := c.Query("active_only") == "true"
    
    filter := bson.M{}
    if activeOnly {
        filter["is_active"] = true
    }
    
    collection := h.db.GetCollection("categories")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    findOptions := options.Find()
    findOptions.SetSort(bson.D{{Key: "sort_order", Value: 1}, {Key: "name", Value: 1}})

    cursor, err := collection.Find(ctx, filter, findOptions)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to fetch categories",
            Error:   stringPtr(err.Error()),
        })
        return
    }
    defer cursor.Close(ctx)

    var categories []models.MongoCategory
    if err = cursor.All(ctx, &categories); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to parse categories",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Categories retrieved successfully",
        Data:    categories,
    })
}

// GetProductsByCategory retrieves all products in a specific category
func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
    categoryID := c.Param("id")
    availableOnly := c.Query("available_only") == "true"

    filter := bson.M{"category_id": categoryID}
    if availableOnly {
        filter["is_available"] = true
    }

    collection := h.db.GetCollection("products")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    findOptions := options.Find()
    findOptions.SetSort(bson.D{{Key: "sort_order", Value: 1}, {Key: "name", Value: 1}})

    cursor, err := collection.Find(ctx, filter, findOptions)
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

func stringPtr(s string) *string {
    return &s
}