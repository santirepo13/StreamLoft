-- Update a user's display name or bitrate
-- Usage: run the query below, edit the WHERE and SET values, then execute

-- Examples:
-- 1. Update name by numeric_id
UPDATE users SET name = 'John Updated', updated_at = NOW() WHERE numeric_id = '100';

-- 2. Update bitrate by user_id (user's configured upload speed in kbps)
UPDATE users SET bitrate = 6000, updated_at = NOW() WHERE id = 1;

-- 3. Update both name and bitrate by numeric_id
UPDATE users SET name = 'Jane Doe', bitrate = 8000, updated_at = NOW() WHERE numeric_id = '200';