-- Regenerate a user's StreamLoft stream key
-- Usage: run the query below, edit the WHERE value, then execute

-- Examples:
-- 1. Regenerate by numeric_id
UPDATE users SET stream_key = SUBSTR(md5(NOW()::text || random()::text), 1, 32), updated_at = NOW() WHERE numeric_id = '1040031189';

-- 2. Regenerate by user_id
UPDATE users SET stream_key = SUBSTR(md5(NOW()::text || random()::text), 1, 32), updated_at = NOW() WHERE id = 1;