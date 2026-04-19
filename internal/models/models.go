package models

import (
    "time"
)

// User represents a system user/staff member
type User struct {
    ID           string    `json:"id" bson:"_id,omitempty"`
    Username     string    `json:"username" bson:"username"`
    Email        string    `json:"email" bson:"email"`
    PasswordHash string    `json:"-" bson:"password_hash"`
    FirstName    string    `json:"first_name" bson:"first_name"`
    LastName     string    `json:"last_name" bson:"last_name"`
    Role         string    `json:"role" bson:"role"`
    IsActive     bool      `json:"is_active" bson:"is_active"`
    CreatedAt    time.Time `json:"created_at" bson:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
}

// Category represents a product category
type Category struct {
    ID          string     `json:"id" bson:"_id,omitempty"`
    Name        string     `json:"name" bson:"name"`
    Description *string    `json:"description" bson:"description"`
    Color       *string    `json:"color" bson:"color"`
    SortOrder   int        `json:"sort_order" bson:"sort_order"`
    IsActive    bool       `json:"is_active" bson:"is_active"`
    CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at" bson:"updated_at"`
}

// Product represents a menu item/product
type Product struct {
    ID              string     `json:"id" bson:"_id,omitempty"`
    CategoryID      *string    `json:"category_id" bson:"category_id,omitempty"`
    Name            string     `json:"name" bson:"name"`
    Description     *string    `json:"description" bson:"description"`
    Price           float64    `json:"price" bson:"price"`
    ImageURL        *string    `json:"image_url" bson:"image_url,omitempty"`
    Barcode         *string    `json:"barcode" bson:"barcode,omitempty"`
    SKU             *string    `json:"sku" bson:"sku,omitempty"`
    IsAvailable     bool       `json:"is_available" bson:"is_available"`
    PreparationTime int        `json:"preparation_time" bson:"preparation_time"`
    SortOrder       int        `json:"sort_order" bson:"sort_order"`
    CreatedAt       time.Time  `json:"created_at" bson:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at" bson:"updated_at"`
    Category        *Category  `json:"category,omitempty" bson:"-"`
}

// DiningTable represents a table or dining area
type DiningTable struct {
    ID              string    `json:"id" bson:"_id,omitempty"`
    TableNumber     string    `json:"table_number" bson:"table_number"`
    SeatingCapacity int       `json:"seating_capacity" bson:"seating_capacity"`
    Location        *string   `json:"location" bson:"location,omitempty"`
    IsOccupied      bool      `json:"is_occupied" bson:"is_occupied"`
    CreatedAt       time.Time `json:"created_at" bson:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" bson:"updated_at"`
}

// Order represents a customer order
type Order struct {
    ID             string       `json:"id" bson:"_id,omitempty"`
    OrderNumber    string       `json:"order_number" bson:"order_number"`
    TableID        *string      `json:"table_id" bson:"table_id,omitempty"`
    UserID         *string      `json:"user_id" bson:"user_id,omitempty"`
    CustomerName   *string      `json:"customer_name" bson:"customer_name,omitempty"`
    OrderType      string       `json:"order_type" bson:"order_type"`
    Status         string       `json:"status" bson:"status"`
    Subtotal       float64      `json:"subtotal" bson:"subtotal"`
    TaxAmount      float64      `json:"tax_amount" bson:"tax_amount"`
    DiscountAmount float64      `json:"discount_amount" bson:"discount_amount"`
    TotalAmount    float64      `json:"total_amount" bson:"total_amount"`
    Notes          *string      `json:"notes" bson:"notes,omitempty"`
    CreatedAt      time.Time    `json:"created_at" bson:"created_at"`
    UpdatedAt      time.Time    `json:"updated_at" bson:"updated_at"`
    ServedAt       *time.Time   `json:"served_at" bson:"served_at,omitempty"`
    CompletedAt    *time.Time   `json:"completed_at" bson:"completed_at,omitempty"`
    Table          *DiningTable `json:"table,omitempty" bson:"-"`
    User           *User        `json:"user,omitempty" bson:"-"`
    Items          []OrderItem  `json:"items,omitempty" bson:"-"`
    Payments       []Payment    `json:"payments,omitempty" bson:"-"`
}

// OrderItem represents an item within an order
type OrderItem struct {
    ID                  string   `json:"id" bson:"_id,omitempty"`
    OrderID             string   `json:"order_id" bson:"order_id"`
    ProductID           string   `json:"product_id" bson:"product_id"`
    Quantity            int      `json:"quantity" bson:"quantity"`
    UnitPrice           float64  `json:"unit_price" bson:"unit_price"`
    TotalPrice          float64  `json:"total_price" bson:"total_price"`
    SpecialInstructions *string  `json:"special_instructions" bson:"special_instructions,omitempty"`
    Status              string   `json:"status" bson:"status"`
    CreatedAt           time.Time `json:"created_at" bson:"created_at"`
    UpdatedAt           time.Time `json:"updated_at" bson:"updated_at"`
    Product             *Product `json:"product,omitempty" bson:"-"`
}

// Payment represents a payment transaction
type Payment struct {
    ID              string     `json:"id" bson:"_id,omitempty"`
    OrderID         string     `json:"order_id" bson:"order_id"`
    PaymentMethod   string     `json:"payment_method" bson:"payment_method"`
    Amount          float64    `json:"amount" bson:"amount"`
    ReferenceNumber *string    `json:"reference_number" bson:"reference_number,omitempty"`
    Status          string     `json:"status" bson:"status"`
    ProcessedBy     *string    `json:"processed_by" bson:"processed_by,omitempty"`
    ProcessedAt     *time.Time `json:"processed_at" bson:"processed_at,omitempty"`
    CreatedAt       time.Time  `json:"created_at" bson:"created_at"`
    ProcessedByUser *User      `json:"processed_by_user,omitempty" bson:"-"`
}

// Inventory represents product inventory
type Inventory struct {
    ID              string     `json:"id" bson:"_id,omitempty"`
    ProductID       string     `json:"product_id" bson:"product_id"`
    CurrentStock    int        `json:"current_stock" bson:"current_stock"`
    MinimumStock    int        `json:"minimum_stock" bson:"minimum_stock"`
    MaximumStock    int        `json:"maximum_stock" bson:"maximum_stock"`
    UnitCost        *float64   `json:"unit_cost" bson:"unit_cost,omitempty"`
    LastRestockedAt *time.Time `json:"last_restocked_at" bson:"last_restocked_at,omitempty"`
    CreatedAt       time.Time  `json:"created_at" bson:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at" bson:"updated_at"`
    Product         *Product   `json:"product,omitempty" bson:"-"`
}

// OrderStatusHistory tracks order status changes
type OrderStatusHistory struct {
    ID             string     `json:"id" bson:"_id,omitempty"`
    OrderID        string     `json:"order_id" bson:"order_id"`
    PreviousStatus *string    `json:"previous_status" bson:"previous_status,omitempty"`
    NewStatus      string     `json:"new_status" bson:"new_status"`
    ChangedBy      *string    `json:"changed_by" bson:"changed_by,omitempty"`
    Notes          *string    `json:"notes" bson:"notes,omitempty"`
    CreatedAt      time.Time  `json:"created_at" bson:"created_at"`
    ChangedByUser  *User      `json:"changed_by_user,omitempty" bson:"-"`
}

// Request/Response DTOs

type CreateOrderRequest struct {
    TableID      *string           `json:"table_id"`
    CustomerName *string           `json:"customer_name"`
    OrderType    string            `json:"order_type"`
    Items        []CreateOrderItem `json:"items"`
    Notes        *string           `json:"notes"`
}

type CreateOrderItem struct {
    ProductID           string  `json:"product_id"`
    Quantity            int     `json:"quantity"`
    SpecialInstructions *string `json:"special_instructions"`
}

type UpdateOrderStatusRequest struct {
    Status string  `json:"status"`
    Notes  *string `json:"notes"`
}

type ProcessPaymentRequest struct {
    PaymentMethod   string  `json:"payment_method"`
    Amount          float64 `json:"amount"`
    ReferenceNumber *string `json:"reference_number"`
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}

type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   *string     `json:"error,omitempty"`
}

type PaginatedResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Meta    MetaData    `json:"meta"`
}

type MetaData struct {
    CurrentPage int `json:"current_page"`
    PerPage     int `json:"per_page"`
    Total       int `json:"total"`
    TotalPages  int `json:"total_pages"`
}