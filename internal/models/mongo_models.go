package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

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

type MongoProduct struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    ProductID       string             `bson:"product_id" json:"product_id"`
    CategoryID      string             `bson:"category_id,omitempty" json:"category_id"`
    Name            string             `bson:"name" json:"name"`
    Description     string             `bson:"description,omitempty" json:"description"`
    Price           float64            `bson:"price" json:"price"`
    ImageURL        string             `bson:"image_url,omitempty" json:"image_url"`
    Barcode         string             `bson:"barcode,omitempty" json:"barcode"`
    SKU             string             `bson:"sku,omitempty" json:"sku"`
    IsAvailable     bool               `bson:"is_available" json:"is_available"`
    PreparationTime int                `bson:"preparation_time" json:"preparation_time"`
    SortOrder       int                `bson:"sort_order" json:"sort_order"`
    CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

type MongoCategory struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    CategoryID  string             `bson:"category_id" json:"category_id"`
    Name        string             `bson:"name" json:"name"`
    Description string             `bson:"description,omitempty" json:"description"`
    Color       string             `bson:"color,omitempty" json:"color"`
    SortOrder   int                `bson:"sort_order" json:"sort_order"`
    IsActive    bool               `bson:"is_active" json:"is_active"`
    CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type MongoOrder struct {
    ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    OrderID        string             `bson:"order_id" json:"order_id"`
    OrderNumber    string             `bson:"order_number" json:"order_number"`
    TableID        string             `bson:"table_id,omitempty" json:"table_id"`
    UserID         string             `bson:"user_id,omitempty" json:"user_id"`
    CustomerName   string             `bson:"customer_name,omitempty" json:"customer_name"`
    OrderType      string             `bson:"order_type" json:"order_type"`
    Status         string             `bson:"status" json:"status"`
    Subtotal       float64            `bson:"subtotal" json:"subtotal"`
    TaxAmount      float64            `bson:"tax_amount" json:"tax_amount"`
    DiscountAmount float64            `bson:"discount_amount" json:"discount_amount"`
    TotalAmount    float64            `bson:"total_amount" json:"total_amount"`
    Notes          string             `bson:"notes,omitempty" json:"notes"`
    CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
    ServedAt       *time.Time         `bson:"served_at,omitempty" json:"served_at"`
    CompletedAt    *time.Time         `bson:"completed_at,omitempty" json:"completed_at"`
}

type MongoTable struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    TableID         string             `bson:"table_id" json:"table_id"`
    TableNumber     string             `bson:"table_number" json:"table_number"`
    SeatingCapacity int                `bson:"seating_capacity" json:"seating_capacity"`
    Location        string             `bson:"location,omitempty" json:"location"`
    IsOccupied      bool               `bson:"is_occupied" json:"is_occupied"`
    CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

type MongoPayment struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    PaymentID       string             `bson:"payment_id" json:"payment_id"`
    OrderID         string             `bson:"order_id" json:"order_id"`
    PaymentMethod   string             `bson:"payment_method" json:"payment_method"`
    Amount          float64            `bson:"amount" json:"amount"`
    ReferenceNumber string             `bson:"reference_number,omitempty" json:"reference_number"`
    Status          string             `bson:"status" json:"status"`
    ProcessedBy     string             `bson:"processed_by,omitempty" json:"processed_by"`
    ProcessedAt     *time.Time         `bson:"processed_at,omitempty" json:"processed_at"`
    CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}