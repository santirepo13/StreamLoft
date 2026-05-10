-- Create a new user
-- Usage: run the query below, edit the VALUES, then execute

-- Examples:
-- 1. Create user with numeric_id=100
INSERT INTO users (numeric_id, name, stream_key, created_at, updated_at)
VALUES ('100', 'John Doe', SUBSTR(md5(NOW()::text || random()::text), 1, 32), NOW(), NOW());

-- 2. Create user with numeric_id=200
INSERT INTO users (numeric_id, name, stream_key, created_at, updated_at)
VALUES ('200', 'Jane Smith', SUBSTR(md5(NOW()::text || random()::text), 1, 32), NOW(), NOW());