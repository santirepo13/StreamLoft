-- Set a user's stream key for a destination
-- Pass the stream key as plain text - it is stored as-is
-- Usage: run the query below, edit the stream_key and WHERE values, then execute

-- Examples:
-- 1. Set stream key for user_destination_id=1
UPDATE user_destinations SET stream_key = '119524913.dc7eb70da236e7812c74cf9f0a105ab0998466be7d34eaba5daca8368155b9c3', updated_at = NOW() WHERE id = 2;

-- 2. Set stream key for user_destination_id=2
UPDATE user_destinations SET stream_key = 'livekey_abc123', updated_at = NOW() WHERE id = 2;

-- 3. Set stream key for user_destination_id=3
UPDATE user_destinations SET stream_key = 'fb-live-00000-11111', updated_at = NOW() WHERE id = 3;