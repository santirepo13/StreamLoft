-- Set a user's stream key for a destination
-- Pass the stream key as plain text - it is stored as-is
-- Usage: run the query below, edit the stream_key and WHERE values, then execute

-- Examples:
-- 1. Set stream key for user_destination_id=1
UPDATE user_destinations SET stream_key = 'sk-live-xxxxx-yyyyy-zzzzz', updated_at = NOW() WHERE id = 1;

-- 2. Set stream key for user_destination_id=2
UPDATE user_destinations SET stream_key = 'livekey_abc123', updated_at = NOW() WHERE id = 2;

-- 3. Set stream key for user_destination_id=3
UPDATE user_destinations SET stream_key = 'fb-live-00000-11111', updated_at = NOW() WHERE id = 3;