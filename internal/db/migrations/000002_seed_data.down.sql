DELETE FROM order_items WHERE id BETWEEN 1 AND 5;
DELETE FROM orders WHERE id BETWEEN 1 AND 3;
DELETE FROM items WHERE id BETWEEN 1 AND 10;
DELETE FROM categories WHERE id BETWEEN 1 AND 5;
DELETE FROM menus WHERE id BETWEEN 1 AND 3;
DELETE FROM users WHERE id BETWEEN 1 AND 3;
DELETE FROM merchants WHERE id BETWEEN 1 AND 3;

SELECT setval('merchants_id_seq', COALESCE((SELECT MAX(id) FROM merchants), 1), (SELECT MAX(id) FROM merchants) IS NOT NULL);
SELECT setval('users_id_seq', COALESCE((SELECT MAX(id) FROM users), 1), (SELECT MAX(id) FROM users) IS NOT NULL);
SELECT setval('menus_id_seq', COALESCE((SELECT MAX(id) FROM menus), 1), (SELECT MAX(id) FROM menus) IS NOT NULL);
SELECT setval('categories_id_seq', COALESCE((SELECT MAX(id) FROM categories), 1), (SELECT MAX(id) FROM categories) IS NOT NULL);
SELECT setval('items_id_seq', COALESCE((SELECT MAX(id) FROM items), 1), (SELECT MAX(id) FROM items) IS NOT NULL);
SELECT setval('orders_id_seq', COALESCE((SELECT MAX(id) FROM orders), 1), (SELECT MAX(id) FROM orders) IS NOT NULL);
SELECT setval('order_items_id_seq', COALESCE((SELECT MAX(id) FROM order_items), 1), (SELECT MAX(id) FROM order_items) IS NOT NULL);
