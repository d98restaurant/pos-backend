package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    // MongoDB Atlas connection string
    uri := "mongodb+srv://admin:mantu1996@cluster0.ap02ozs.mongodb.net/?appName=Cluster0"
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        log.Fatal("Failed to connect:", err)
    }
    defer client.Disconnect(ctx)
    
    if err := client.Ping(ctx, nil); err != nil {
        log.Fatal("Failed to ping:", err)
    }
    
    fmt.Println("✅ Connected to MongoDB Atlas!")
    
    db := client.Database("pos_system")
    
    // Create collections
    collections := []string{"users", "categories", "products", "tables", "orders", "order_items", "payments", "inventory", "order_status_history"}
    for _, coll := range collections {
        err := db.CreateCollection(ctx, coll)
        if err != nil {
            // Collection might already exist
            fmt.Printf("Collection %s: %v\n", coll, err)
        } else {
            fmt.Printf("✅ Created collection: %s\n", coll)
        }
    }
    
    // Create indexes
    fmt.Println("\n📇 Creating indexes...")
    
    // Users indexes
    db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys:    bson.D{{Key: "username", Value: 1}},
        Options: options.Index().SetUnique(true),
    })
    db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{Key: "email", Value: 1}},
        Options: options.Index().SetUnique(true),
    })
    db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{Key: "role", Value: 1}},
    })
    
    // Categories indexes
    db.Collection("categories").Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{Key: "name", Value: 1}},
    })
    db.Collection("categories").Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{Key: "sort_order", Value: 1}},
    })
    
    // Products indexes
    db.Collection("products").Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{Key: "category_id", Value: 1}},
    })
    db.Collection("products").Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{Key: "name", Value: 1}},
    })
    db.Collection("products").Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{Key: "is_available", Value: 1}},
    })
    
    // Tables indexes
    db.Collection("tables").Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys:    bson.D{{Key: "table_number", Value: 1}},
        Options: options.Index().SetUnique(true),
    })
    
    // Orders indexes
    db.Collection("orders").Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys:    bson.D{{Key: "order_number", Value: 1}},
        Options: options.Index().SetUnique(true),
    })
    db.Collection("orders").Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{Key: "status", Value: 1}},
    })
    db.Collection("orders").Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{Key: "created_at", Value: -1}},
    })
    
    fmt.Println("✅ Indexes created successfully")
    
    // Seed users (password: admin123)
    fmt.Println("\n👥 Seeding users...")
    users := []interface{}{
        bson.M{
            "username":    "admin",
            "email":       "admin@pos.com",
            "password_hash": "$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO",
            "first_name":  "Admin",
            "last_name":   "User",
            "role":        "admin",
            "is_active":   true,
            "created_at":  time.Now(),
            "updated_at":  time.Now(),
        },
        bson.M{
            "username":    "manager1",
            "email":       "manager@pos.com",
            "password_hash": "$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO",
            "first_name":  "John",
            "last_name":   "Manager",
            "role":        "manager",
            "is_active":   true,
            "created_at":  time.Now(),
            "updated_at":  time.Now(),
        },
        bson.M{
            "username":    "server1",
            "email":       "server1@pos.com",
            "password_hash": "$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO",
            "first_name":  "Sarah",
            "last_name":   "Smith",
            "role":        "server",
            "is_active":   true,
            "created_at":  time.Now(),
            "updated_at":  time.Now(),
        },
        bson.M{
            "username":    "server2",
            "email":       "server2@pos.com",
            "password_hash": "$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO",
            "first_name":  "Mike",
            "last_name":   "Johnson",
            "role":        "server",
            "is_active":   true,
            "created_at":  time.Now(),
            "updated_at":  time.Now(),
        },
        bson.M{
            "username":    "counter1",
            "email":       "counter1@pos.com",
            "password_hash": "$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO",
            "first_name":  "Lisa",
            "last_name":   "Davis",
            "role":        "counter",
            "is_active":   true,
            "created_at":  time.Now(),
            "updated_at":  time.Now(),
        },
        bson.M{
            "username":    "counter2",
            "email":       "counter2@pos.com",
            "password_hash": "$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO",
            "first_name":  "Tom",
            "last_name":   "Wilson",
            "role":        "counter",
            "is_active":   true,
            "created_at":  time.Now(),
            "updated_at":  time.Now(),
        },
        bson.M{
            "username":    "kitchen1",
            "email":       "kitchen@pos.com",
            "password_hash": "$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO",
            "first_name":  "Chef",
            "last_name":   "Williams",
            "role":        "kitchen",
            "is_active":   true,
            "created_at":  time.Now(),
            "updated_at":  time.Now(),
        },
    }
    
    for _, user := range users {
        opts := options.Update().SetUpsert(true)
        filter := bson.M{"username": user.(bson.M)["username"]}
        _, err := db.Collection("users").UpdateOne(ctx, filter, bson.M{"$setOnInsert": user}, opts)
        if err != nil {
            log.Printf("Error inserting user: %v", err)
        }
    }
    fmt.Println("✅ Users seeded: 7 users (password: admin123)")
    
    // Seed categories
    fmt.Println("\n🏷️ Seeding categories...")
    categories := []interface{}{
        bson.M{"name": "Appetizers", "description": "Starter dishes and small plates", "color": "#FF6B6B", "sort_order": 1, "is_active": true, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"name": "Main Courses", "description": "Primary dishes and entrees", "color": "#4ECDC4", "sort_order": 2, "is_active": true, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"name": "Beverages", "description": "Drinks, sodas, and refreshments", "color": "#45B7D1", "sort_order": 3, "is_active": true, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"name": "Desserts", "description": "Sweet treats and desserts", "color": "#96CEB4", "sort_order": 4, "is_active": true, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"name": "Salads", "description": "Fresh salads and healthy options", "color": "#FECA57", "sort_order": 5, "is_active": true, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"name": "Pizza", "description": "Various pizza options", "color": "#FF9FF3", "sort_order": 6, "is_active": true, "created_at": time.Now(), "updated_at": time.Now()},
    }
    
    for _, cat := range categories {
        opts := options.Update().SetUpsert(true)
        filter := bson.M{"name": cat.(bson.M)["name"]}
        _, err := db.Collection("categories").UpdateOne(ctx, filter, bson.M{"$setOnInsert": cat}, opts)
        if err != nil {
            log.Printf("Error inserting category: %v", err)
        }
    }
    fmt.Println("✅ Categories seeded: 6 categories")
    
    // Get category IDs
    var categoryDocs []bson.M
    cursor, _ := db.Collection("categories").Find(ctx, bson.M{})
    cursor.All(ctx, &categoryDocs)
    
    categoryMap := make(map[string]primitive.ObjectID)
    for _, cat := range categoryDocs {
        categoryMap[cat["name"].(string)] = cat["_id"].(primitive.ObjectID)
    }
    
    // Seed products
    fmt.Println("\n🍕 Seeding products...")
    products := []interface{}{
        bson.M{"name": "Garlic Bread", "description": "Toasted bread with garlic butter and herbs", "price": 4.99, "category_id": categoryMap["Appetizers"], "is_available": true, "preparation_time": 8, "sort_order": 1, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"name": "Chicken Wings", "description": "Spicy buffalo wings with ranch dip", "price": 8.99, "category_id": categoryMap["Appetizers"], "is_available": true, "preparation_time": 12, "sort_order": 2, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"name": "Margherita Pizza", "description": "Fresh mozzarella, tomato sauce, basil", "price": 12.99, "category_id": categoryMap["Pizza"], "is_available": true, "preparation_time": 15, "sort_order": 1, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"name": "Pepperoni Pizza", "description": "Classic pepperoni with mozzarella", "price": 14.99, "category_id": categoryMap["Pizza"], "is_available": true, "preparation_time": 15, "sort_order": 2, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"name": "Grilled Chicken", "description": "Grilled chicken breast with vegetables", "price": 16.99, "category_id": categoryMap["Main Courses"], "is_available": true, "preparation_time": 20, "sort_order": 1, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"name": "Caesar Salad", "description": "Romaine lettuce, croutons, parmesan", "price": 7.99, "category_id": categoryMap["Salads"], "is_available": true, "preparation_time": 5, "sort_order": 1, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"name": "Coke", "description": "Refreshing cola drink", "price": 2.49, "category_id": categoryMap["Beverages"], "is_available": true, "preparation_time": 1, "sort_order": 1, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"name": "Chocolate Cake", "description": "Rich chocolate layer cake", "price": 5.99, "category_id": categoryMap["Desserts"], "is_available": true, "preparation_time": 3, "sort_order": 1, "created_at": time.Now(), "updated_at": time.Now()},
    }
    
    for _, prod := range products {
        opts := options.Update().SetUpsert(true)
        filter := bson.M{"name": prod.(bson.M)["name"]}
        _, err := db.Collection("products").UpdateOne(ctx, filter, bson.M{"$setOnInsert": prod}, opts)
        if err != nil {
            log.Printf("Error inserting product: %v", err)
        }
    }
    fmt.Println("✅ Products seeded: 8 products")
    
    // Seed tables
    fmt.Println("\n🪑 Seeding tables...")
    tables := []interface{}{
        bson.M{"table_number": "T01", "seating_capacity": 2, "location": "Main Floor", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "T02", "seating_capacity": 4, "location": "Main Floor", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "T03", "seating_capacity": 4, "location": "Main Floor", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "T04", "seating_capacity": 6, "location": "Main Floor", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "T05", "seating_capacity": 2, "location": "Main Floor", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "T06", "seating_capacity": 4, "location": "Window Side", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "T07", "seating_capacity": 4, "location": "Window Side", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "T08", "seating_capacity": 8, "location": "Private Room", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "T09", "seating_capacity": 2, "location": "Patio", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "T10", "seating_capacity": 4, "location": "Patio", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "BAR01", "seating_capacity": 1, "location": "Bar Counter", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "BAR02", "seating_capacity": 1, "location": "Bar Counter", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "BAR03", "seating_capacity": 1, "location": "Bar Counter", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
        bson.M{"table_number": "TAKEOUT", "seating_capacity": 1, "location": "Takeout Counter", "is_occupied": false, "created_at": time.Now(), "updated_at": time.Now()},
    }
    
    for _, table := range tables {
        opts := options.Update().SetUpsert(true)
        filter := bson.M{"table_number": table.(bson.M)["table_number"]}
        _, err := db.Collection("tables").UpdateOne(ctx, filter, bson.M{"$setOnInsert": table}, opts)
        if err != nil {
            log.Printf("Error inserting table: %v", err)
        }
    }
    fmt.Println("✅ Tables seeded: 14 tables")
    
    fmt.Println("\n🎉 Database initialization completed successfully!")
    fmt.Println("🔐 Login credentials (password: admin123 for all users):")
    fmt.Println("   - admin (Admin)")
    fmt.Println("   - manager1 (Manager)")
    fmt.Println("   - server1 (Server)")
    fmt.Println("   - server2 (Server)")
    fmt.Println("   - counter1 (Counter)")
    fmt.Println("   - counter2 (Counter)")
    fmt.Println("   - kitchen1 (Kitchen)")
}