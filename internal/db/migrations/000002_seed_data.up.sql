INSERT INTO merchants (id, name, status) VALUES
    (1, 'Joe''s Pizza', 'ACTIVE'),
    (2, 'Cafe Bloom', 'ACTIVE'),
    (3, 'Sushi Zen', 'INACTIVE');

INSERT INTO users (id, name, email) VALUES
    (1, 'Alice', 'alice@example.com'),
    (2, 'Bob', 'bob@example.com'),
    (3, 'Carol', 'carol@example.com');

INSERT INTO menus (id, merchant_id, name, status) VALUES
    (1, 1, 'Main Menu', 'ACTIVE'),
    (2, 1, 'Winter Specials', 'INACTIVE'),
    (3, 2, 'All Day Menu', 'ACTIVE');

INSERT INTO categories (id, name, menu_id, status) VALUES
    (1, 'Pizzas', 1, 'ACTIVE'),
    (2, 'Sides', 1, 'ACTIVE'),
    (3, 'Drinks', 1, 'ACTIVE'),
    (4, 'Coffee', 3, 'ACTIVE'),
    (5, 'Pastries', 3, 'ACTIVE');

INSERT INTO items (id, merchant_id, name, price, status, availability, category_id) VALUES
    (1, 1, 'Margherita Pizza', 12.99, 'ACTIVE', 'AVAILABLE', 1),
    (2, 1, 'Pepperoni Pizza', 14.99, 'ACTIVE', 'AVAILABLE', 1),
    (3, 1, 'Hawaiian Pizza', 13.99, 'INACTIVE', 'AVAILABLE', 1),
    (4, 1, 'Garlic Bread', 5.99, 'ACTIVE', 'AVAILABLE', 2),
    (5, 1, 'Coke', 2.50, 'ACTIVE', 'OUT_OF_STOCK', 3),
    (6, 1, 'Iced Tea', 3.00, 'ACTIVE', 'AVAILABLE', 3),
    (7, 2, 'Latte', 4.50, 'ACTIVE', 'AVAILABLE', 4),
    (8, 2, 'Espresso', 3.50, 'ACTIVE', 'AVAILABLE', 4),
    (9, 2, 'Croissant', 3.99, 'ACTIVE', 'AVAILABLE', 5),
    (10, 2, 'Blueberry Muffin', 2.99, 'ACTIVE', 'AVAILABLE', 5);

INSERT INTO orders (id, user_id, merchant_id, status, total_amount) VALUES
    (1, 1, 1, 'COMPLETED', 31.97),
    (2, 2, 1, 'PENDING', 14.99),
    (3, 3, 2, 'CONFIRMED', 12.48);

INSERT INTO order_items (id, order_id, item_id, quantity, unit_price) VALUES
    (1, 1, 1, 2, 12.99),
    (2, 1, 4, 1, 5.99),
    (3, 2, 2, 1, 14.99),
    (4, 3, 7, 1, 4.50),
    (5, 3, 9, 2, 3.99);

SELECT setval('merchants_id_seq', (SELECT MAX(id) FROM merchants));
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
SELECT setval('menus_id_seq', (SELECT MAX(id) FROM menus));
SELECT setval('categories_id_seq', (SELECT MAX(id) FROM categories));
SELECT setval('items_id_seq', (SELECT MAX(id) FROM items));
SELECT setval('orders_id_seq', (SELECT MAX(id) FROM orders));
SELECT setval('order_items_id_seq', (SELECT MAX(id) FROM order_items));
