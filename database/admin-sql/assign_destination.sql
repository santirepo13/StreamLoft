-- Assign a destination to a user
-- The destination must first exist in the destinations table
-- Usage: run the query below, edit the VALUES, then execute

-- Examples:
-- 1. Assign YouTube to user with numeric_id=100 (by destination name)
INSERT INTO user_destinations (user_id, destination_id, created_at, updated_at)
SELECT u.id, d.id, NOW(), NOW()
FROM users u
CROSS JOIN destinations d
WHERE u.numeric_id = '100' AND d.name = 'YouTube';

-- 2. Assign Twitch to user with numeric_id=200 (by destination id=2)
INSERT INTO user_destinations (user_id, destination_id, created_at, updated_at)
SELECT u.id, d.id, NOW(), NOW()
FROM users u
CROSS JOIN destinations d
WHERE u.numeric_id = '200' AND d.id = 2;