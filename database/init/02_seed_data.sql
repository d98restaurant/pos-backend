-- Seed data for POS System

-- Insert default users (password is 'admin123' hashed)
INSERT INTO users (username, email, password_hash, first_name, last_name, role) VALUES
('admin', 'admin@pos.com', '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO', 'Admin', 'User', 'admin'),
('manager1', 'manager@pos.com', '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO', 'John', 'Manager', 'manager'),
('server1', 'server1@pos.com', '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO', 'Sarah', 'Smith', 'server'),
('server2', 'server2@pos.com', '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO', 'Mike', 'Johnson', 'server'),
('counter1', 'counter1@pos.com', '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO', 'Lisa', 'Davis', 'counter'),
('counter2', 'counter2@pos.com', '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO', 'Tom', 'Wilson', 'counter'),
('kitchen1', 'kitchen@pos.com', '$2a$10$FPH.ONfAgquWmXjM3LE61OIgOPgXX8i.jOISCHZ2DpK2gg4krEWfO', 'Chef', 'Williams', 'kitchen')
ON CONFLICT (username) DO NOTHING;

-- Insert categories
INSERT INTO categories (name, description, color, sort_order) VALUES
('Appetizers', 'Starter dishes and small plates', '#FF6B6B', 1),
('Main Courses', 'Primary dishes and entrees', '#4ECDC4', 2),
('Beverages', 'Drinks, sodas, and refreshments', '#45B7D1', 3),
('Desserts', 'Sweet treats and desserts', '#96CEB4', 4),
('Salads', 'Fresh salads and healthy options', '#FECA57', 5),
('Pizza', 'Various pizza options', '#FF9FF3', 6)
ON CONFLICT (name) DO NOTHING;

-- Insert dining tables
INSERT INTO dining_tables (table_number, seating_capacity, location) VALUES
('T01', 2, 'Main Floor'),
('T02', 4, 'Main Floor'),
('T03', 4, 'Main Floor'),
('T04', 6, 'Main Floor'),
('T05', 2, 'Main Floor'),
('T06', 4, 'Window Side'),
('T07', 4, 'Window Side'),
('T08', 8, 'Private Room'),
('T09', 2, 'Patio'),
('T10', 4, 'Patio'),
('BAR01', 1, 'Bar Counter'),
('BAR02', 1, 'Bar Counter'),
('BAR03', 1, 'Bar Counter'),
('TAKEOUT', 1, 'Takeout Counter')
ON CONFLICT (table_number) DO NOTHING;
