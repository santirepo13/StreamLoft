-- Update a user's display name
-- Usage: run the query below, edit the WHERE and SET values, then execute

-- Examples:
-- 1. Update by numeric_id
UPDATE users SET name = 'John Updated', updated_at = NOW() WHERE numeric_id = '100';

-- 2. Update by user_id
UPDATE users SET name = 'Jane Doe', updated_at = NOW() WHERE id = 1;