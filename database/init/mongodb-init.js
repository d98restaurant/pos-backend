// MongoDB initialization script for POS system
// Run this to create indexes and seed initial data

// Switch to pos_system database
db = db.getSiblingDB('pos_system');

// ==================== CREATE COLLECTIONS ====================
db.createCollection('users');
db.createCollection('categories');
db.createCollection('products');
db.createCollection('tables');
db.createCollection('orders');
db.createCollection('order_items');
db.createCollection('payments');
db.createCollection('inventory');
db.createCollection('order_status_history');

// ==================== CREATE INDEXES ====================

// Users indexes
db.users.createIndex({ username: 1 }, { unique: true });
db.users.createIndex({ email: 1 }, { unique: true });
db.users.createIndex({ role: 1 });

// Categories indexes
db.categories.createIndex({ name: 1 });
db.categories.createIndex({ sort_order: 1 });

// Products indexes
db.products.createIndex({ category_id: 1 });
db.products.createIndex({ name: 1 });
db.products.createIndex({ is_available: 1 });
db.products.createIndex({ sku: 1 }, { unique: true, sparse: true });

// Tables indexes
db.tables.createIndex({ table_number: 1 }, { unique: true });
db.tables.createIndex({ is_occupied: 1 });
db.tables.createIndex({ location: 1 });

// Orders indexes
db.orders.createIndex({ order_number: 1 }, { unique: true });
db.orders.createIndex({ status: 1 });
db.orders.createIndex({ created_at: -1 });
db.orders.createIndex({ table_id: 1 });
db.orders.createIndex({ order_type: 1 });
db.orders.createIndex({ user_id: 1 });

// Order items indexes
db.order_items.createIndex({ order_id: 1 });
db.order_items.createIndex({ product_id: 1 });
db.order_items.createIndex({ status: 1 });

// Payments indexes
db.payments.createIndex({ order_id: 1 });
db.payments.createIndex({ status: 1 });
db.payments.createIndex({ processed_at: -1 });

// Inventory indexes
db.inventory.createIndex({ product_id: 1 }, { unique: true });

// Order status history indexes
db.order_status_history.createIndex({ order_id: 1 });
db.order_status_history.createIndex({ created_at: -1 });

// ==================== SEED DATA ====================

// Seed users (password is 'admin123' hashed with bcrypt)
// The hash below is for 'admin123' - you may need to regenerate
db.users.insertMany([
  {
    username: 'admin',
    email: 'admin@pos.com',
    password_hash: '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO',
    first_name: 'Admin',
    last_name: 'User',
    role: 'admin',
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    username: 'manager1',
    email: 'manager@pos.com',
    password_hash: '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO',
    first_name: 'John',
    last_name: 'Manager',
    role: 'manager',
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    username: 'server1',
    email: 'server1@pos.com',
    password_hash: '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO',
    first_name: 'Sarah',
    last_name: 'Smith',
    role: 'server',
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    username: 'server2',
    email: 'server2@pos.com',
    password_hash: '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO',
    first_name: 'Mike',
    last_name: 'Johnson',
    role: 'server',
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    username: 'counter1',
    email: 'counter1@pos.com',
    password_hash: '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO',
    first_name: 'Lisa',
    last_name: 'Davis',
    role: 'counter',
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    username: 'counter2',
    email: 'counter2@pos.com',
    password_hash: '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO',
    first_name: 'Tom',
    last_name: 'Wilson',
    role: 'counter',
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    username: 'kitchen1',
    email: 'kitchen@pos.com',
    password_hash: '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO',
    first_name: 'Chef',
    last_name: 'Williams',
    role: 'kitchen',
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  }
]);

// Seed categories
db.categories.insertMany([
  {
    name: 'Appetizers',
    description: 'Starter dishes and small plates',
    color: '#FF6B6B',
    sort_order: 1,
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Main Courses',
    description: 'Primary dishes and entrees',
    color: '#4ECDC4',
    sort_order: 2,
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Beverages',
    description: 'Drinks, sodas, and refreshments',
    color: '#45B7D1',
    sort_order: 3,
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Desserts',
    description: 'Sweet treats and desserts',
    color: '#96CEB4',
    sort_order: 4,
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Salads',
    description: 'Fresh salads and healthy options',
    color: '#FECA57',
    sort_order: 5,
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Pizza',
    description: 'Various pizza options',
    color: '#FF9FF3',
    sort_order: 6,
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  }
]);

// Get category IDs for reference
var categories = db.categories.find().toArray();
var categoryMap = {};
categories.forEach(function(cat) {
  categoryMap[cat.name] = cat._id;
});

// Seed products
db.products.insertMany([
  {
    name: 'Garlic Bread',
    description: 'Toasted bread with garlic butter and herbs',
    price: 4.99,
    category_id: categoryMap['Appetizers'],
    is_available: true,
    preparation_time: 8,
    sort_order: 1,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Chicken Wings',
    description: 'Spicy buffalo wings with ranch dip',
    price: 8.99,
    category_id: categoryMap['Appetizers'],
    is_available: true,
    preparation_time: 12,
    sort_order: 2,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Margherita Pizza',
    description: 'Fresh mozzarella, tomato sauce, basil',
    price: 12.99,
    category_id: categoryMap['Pizza'],
    is_available: true,
    preparation_time: 15,
    sort_order: 1,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Pepperoni Pizza',
    description: 'Classic pepperoni with mozzarella',
    price: 14.99,
    category_id: categoryMap['Pizza'],
    is_available: true,
    preparation_time: 15,
    sort_order: 2,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Grilled Chicken',
    description: 'Grilled chicken breast with vegetables',
    price: 16.99,
    category_id: categoryMap['Main Courses'],
    is_available: true,
    preparation_time: 20,
    sort_order: 1,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Caesar Salad',
    description: 'Romaine lettuce, croutons, parmesan',
    price: 7.99,
    category_id: categoryMap['Salads'],
    is_available: true,
    preparation_time: 5,
    sort_order: 1,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Coke',
    description: 'Refreshing cola drink',
    price: 2.49,
    category_id: categoryMap['Beverages'],
    is_available: true,
    preparation_time: 1,
    sort_order: 1,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Chocolate Cake',
    description: 'Rich chocolate layer cake',
    price: 5.99,
    category_id: categoryMap['Desserts'],
    is_available: true,
    preparation_time: 3,
    sort_order: 1,
    created_at: new Date(),
    updated_at: new Date()
  }
]);

// Seed dining tables
db.tables.insertMany([
  { table_number: 'T01', seating_capacity: 2, location: 'Main Floor', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'T02', seating_capacity: 4, location: 'Main Floor', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'T03', seating_capacity: 4, location: 'Main Floor', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'T04', seating_capacity: 6, location: 'Main Floor', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'T05', seating_capacity: 2, location: 'Main Floor', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'T06', seating_capacity: 4, location: 'Window Side', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'T07', seating_capacity: 4, location: 'Window Side', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'T08', seating_capacity: 8, location: 'Private Room', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'T09', seating_capacity: 2, location: 'Patio', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'T10', seating_capacity: 4, location: 'Patio', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'BAR01', seating_capacity: 1, location: 'Bar Counter', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'BAR02', seating_capacity: 1, location: 'Bar Counter', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'BAR03', seating_capacity: 1, location: 'Bar Counter', is_occupied: false, created_at: new Date(), updated_at: new Date() },
  { table_number: 'TAKEOUT', seating_capacity: 1, location: 'Takeout Counter', is_occupied: false, created_at: new Date(), updated_at: new Date() }
]);

print('MongoDB initialization completed!');
print('Collections created: users, categories, products, tables, orders, order_items, payments, inventory, order_status_history');
print('Indexes created successfully');
print('Seed data inserted: 7 users, 6 categories, 8 products, 14 tables');