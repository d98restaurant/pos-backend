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

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("products")
    total, _ := collection.CountDocuments(ctx, filter)

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

    var products []bson.M
    if err = cursor.All(ctx, &products); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to scan product",
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

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("products")
    var product bson.M
    err := collection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product)

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

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("categories")
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

    var categories []bson.M
    if err = cursor.All(ctx, &categories); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to scan category",
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

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("products")
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

    var products []bson.M
    if err = cursor.All(ctx, &products); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to scan product",
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

// GetAdminCategories - Admin endpoint for categories
func (h *ProductHandler) GetAdminCategories(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("categories")
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to fetch categories",
            Error:   stringPtr(err.Error()),
        })
        return
    }
    defer cursor.Close(ctx)

    var categories []bson.M
    cursor.All(ctx, &categories)

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Categories retrieved successfully",
        Data:    categories,
    })
}

// GetAdminProducts - Admin endpoint for products
func (h *ProductHandler) GetAdminProducts(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("products")
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to fetch products",
            Error:   stringPtr(err.Error()),
        })
        return
    }
    defer cursor.Close(ctx)

    var products []bson.M
    cursor.All(ctx, &products)

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Products retrieved successfully",
        Data:    products,
    })
}

// CreateCategory creates a new category
func (h *ProductHandler) CreateCategory(c *gin.Context) {
    var req struct {
        Name        string  `json:"name" binding:"required"`
        Description *string `json:"description"`
        Color       *string `json:"color"`
        SortOrder   int     `json:"sort_order"`
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

    collection := h.db.GetCollection("categories")
    result, err := collection.InsertOne(ctx, bson.M{
        "name":        req.Name,
        "description": req.Description,
        "color":       req.Color,
        "sort_order":  req.SortOrder,
        "is_active":   true,
        "created_at":  time.Now(),
        "updated_at":  time.Now(),
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to create category",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    c.JSON(http.StatusCreated, models.APIResponse{
        Success: true,
        Message: "Category created successfully",
        Data:    map[string]interface{}{"id": result.InsertedID},
    })
}

// UpdateCategory updates an existing category
func (h *ProductHandler) UpdateCategory(c *gin.Context) {
    categoryID := c.Param("id")

    var req struct {
        Name        *string `json:"name"`
        Description *string `json:"description"`
        Color       *string `json:"color"`
        SortOrder   *int    `json:"sort_order"`
        IsActive    *bool   `json:"is_active"`
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
    if req.Name != nil {
        update["$set"].(bson.M)["name"] = *req.Name
    }
    if req.Description != nil {
        update["$set"].(bson.M)["description"] = *req.Description
    }
    if req.Color != nil {
        update["$set"].(bson.M)["color"] = *req.Color
    }
    if req.SortOrder != nil {
        update["$set"].(bson.M)["sort_order"] = *req.SortOrder
    }
    if req.IsActive != nil {
        update["$set"].(bson.M)["is_active"] = *req.IsActive
    }

    collection := h.db.GetCollection("categories")
    result, err := collection.UpdateOne(ctx, bson.M{"_id": categoryID}, update)

    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to update category",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Category not found",
            Error:   stringPtr("category_not_found"),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Category updated successfully",
    })
}

// DeleteCategory deletes a category
func (h *ProductHandler) DeleteCategory(c *gin.Context) {
    categoryID := c.Param("id")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("categories")
    result, err := collection.DeleteOne(ctx, bson.M{"_id": categoryID})

    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to delete category",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    if result.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Category not found",
            Error:   stringPtr("category_not_found"),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Category deleted successfully",
    })
}

// CreateProduct creates a new product
func (h *ProductHandler) CreateProduct(c *gin.Context) {
    var req struct {
        CategoryID      *string `json:"category_id"`
        Name            string  `json:"name" binding:"required"`
        Description     *string `json:"description"`
        Price           float64 `json:"price" binding:"required"`
        ImageURL        *string `json:"image_url"`
        Barcode         *string `json:"barcode"`
        SKU             *string `json:"sku"`
        PreparationTime int     `json:"preparation_time"`
        SortOrder       int     `json:"sort_order"`
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

    collection := h.db.GetCollection("products")
    product := bson.M{
        "name":             req.Name,
        "description":      req.Description,
        "price":            req.Price,
        "image_url":        req.ImageURL,
        "barcode":          req.Barcode,
        "sku":              req.SKU,
        "preparation_time": req.PreparationTime,
        "sort_order":       req.SortOrder,
        "is_available":     true,
        "created_at":       time.Now(),
        "updated_at":       time.Now(),
    }
    if req.CategoryID != nil {
        product["category_id"] = *req.CategoryID
    }

    result, err := collection.InsertOne(ctx, product)

    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to create product",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    c.JSON(http.StatusCreated, models.APIResponse{
        Success: true,
        Message: "Product created successfully",
        Data:    map[string]interface{}{"id": result.InsertedID},
    })
}

// UpdateProduct updates an existing product
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
    productID := c.Param("id")

    var req struct {
        CategoryID      *string  `json:"category_id"`
        Name            *string  `json:"name"`
        Description     *string  `json:"description"`
        Price           *float64 `json:"price"`
        ImageURL        *string  `json:"image_url"`
        Barcode         *string  `json:"barcode"`
        SKU             *string  `json:"sku"`
        IsAvailable     *bool    `json:"is_available"`
        PreparationTime *int     `json:"preparation_time"`
        SortOrder       *int     `json:"sort_order"`
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
    if req.Name != nil {
        update["$set"].(bson.M)["name"] = *req.Name
    }
    if req.Description != nil {
        update["$set"].(bson.M)["description"] = *req.Description
    }
    if req.Price != nil {
        update["$set"].(bson.M)["price"] = *req.Price
    }
    if req.ImageURL != nil {
        update["$set"].(bson.M)["image_url"] = *req.ImageURL
    }
    if req.Barcode != nil {
        update["$set"].(bson.M)["barcode"] = *req.Barcode
    }
    if req.SKU != nil {
        update["$set"].(bson.M)["sku"] = *req.SKU
    }
    if req.IsAvailable != nil {
        update["$set"].(bson.M)["is_available"] = *req.IsAvailable
    }
    if req.PreparationTime != nil {
        update["$set"].(bson.M)["preparation_time"] = *req.PreparationTime
    }
    if req.SortOrder != nil {
        update["$set"].(bson.M)["sort_order"] = *req.SortOrder
    }
    if req.CategoryID != nil {
        update["$set"].(bson.M)["category_id"] = *req.CategoryID
    }

    collection := h.db.GetCollection("products")
    result, err := collection.UpdateOne(ctx, bson.M{"_id": productID}, update)

    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to update product",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Product not found",
            Error:   stringPtr("product_not_found"),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Product updated successfully",
    })
}

// DeleteProduct deletes a product
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
    productID := c.Param("id")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := h.db.GetCollection("products")
    result, err := collection.DeleteOne(ctx, bson.M{"_id": productID})

    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: "Failed to delete product",
            Error:   stringPtr(err.Error()),
        })
        return
    }

    if result.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "Product not found",
            Error:   stringPtr("product_not_found"),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Product deleted successfully",
    })
}
