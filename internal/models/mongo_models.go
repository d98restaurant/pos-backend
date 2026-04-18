package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// MongoUser represents a system user in MongoDB
type MongoUser struct {
    ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID       string             `bson:"user_id" json:"user_id"`
    Username     string             `bson:"username" json:"username"`
    Email        string             `bson:"email" json:"email"`
    PasswordHash string             `bson:"password_hash" json:"-"`
    FirstName    string             `bson:"first_name" json:"first_name"`
    LastName     string             `bson:"last_name" json:"last_name"`
    Role         string             `bson:"role" json:"role"`
    IsActive     bool               `bson:"is_active" json:"is_active"`
    CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// MongoProduct represents a product in MongoDB
type MongoProduct struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    ProductID       string             `bson:"product_id" json:"product_id"`
    CategoryID      *string            `bson:"category_id,omitempty" json:"category_id"`
    Name            string             `bson:"name" json:"name"`
    Description     *string            `bson:"description,omitempty" json:"description"`
    Price           float64            `bson:"price" json:"price"`
    ImageURL        *string            `bson:"image_url,omitempty" json:"image_url"`
    Barcode         *string            `bson:"barcode,omitempty" json:"barcode"`
    SKU             *string            `bson:"sku,omitempty" json:"sku"`
    IsAvailable     bool               `bson:"is_available" json:"is_available"`
    PreparationTime int                `bson:"preparation_time" json:"preparation_time"`
    SortOrder       int                `bson:"sort_order" json:"sort_order"`
    CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

// MongoOrder represents an order in MongoDB
type MongoOrder struct {
    ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    OrderID        string             `bson:"order_id" json:"order_id"`
    OrderNumber    string             `bson:"order_number" json:"order_number"`
    TableID        *string            `bson:"table_id,omitempty" json:"table_id"`
    UserID         *string            `bson:"user_id,omitempty" json:"user_id"`
    CustomerName   *string            `bson:"customer_name,omitempty" json:"customer_name"`
    OrderType      string             `bson:"order_type" json:"order_type"`
    Status         string             `bson:"status" json:"status"`
    Subtotal       float64            `bson:"subtotal" json:"subtotal"`
    TaxAmount      float64            `bson:"tax_amount" json:"tax_amount"`
    DiscountAmount float64            `bson:"discount_amount" json:"discount_amount"`
    TotalAmount    float64            `bson:"total_amount" json:"total_amount"`
    Notes          *string            `bson:"notes,omitempty" json:"notes"`
    CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
    ServedAt       *time.Time         `bson:"served_at,omitempty" json:"served_at"`
    CompletedAt    *time.Time         `bson:"completed_at,omitempty" json:"completed_at"`
}