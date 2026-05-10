-- List destinations for a user
-- Usage: run the query below, edit the WHERE value, then execute

-- Examples:
-- 1. List destinations for user with numeric_id=100
SELECT
    ud.id AS user_destination_id,
    d.id AS destination_id,
    u.id AS user_id,
    u.numeric_id,
    u.name AS user_name,
    d.name AS destination_name,
    d.rtmp_url,
    ud.stream_key,
    ud.created_at,
    ud.updated_at
FROM user_destinations ud
JOIN destinations d ON d.id = ud.destination_id
JOIN users u ON u.id = ud.user_id
WHERE u.numeric_id = '100'
ORDER BY d.name;

-- 2. List destinations for user with user_id=1
SELECT
    ud.id AS user_destination_id,
    d.id AS destination_id,
    u.id AS user_id,
    u.numeric_id,
    u.name AS user_name,
    d.name AS destination_name,
    d.rtmp_url,
    ud.stream_key,
    ud.created_at,
    ud.updated_at
FROM user_destinations ud
JOIN destinations d ON d.id = ud.destination_id
JOIN users u ON u.id = ud.user_id
WHERE u.id = 1
ORDER BY d.name;