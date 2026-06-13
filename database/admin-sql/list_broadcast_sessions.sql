-- List broadcast sessions
-- Usage: run the query below, edit the WHERE values, then execute

-- Examples:
-- 1. List sessions for destination id=1 (all users)
SELECT
    bs.id AS session_id,
    u.id AS user_id,
    u.numeric_id,
    u.name AS user_name,
    d.id AS destination_id,
    d.name AS destination_name,
    bs.date,
    bs.duration_minutes,
    bs.started_at,
    bs.ended_at
FROM broadcast_sessions bs
JOIN user_destinations ud ON ud.id = bs.user_destination_id
JOIN destinations d ON d.id = ud.destination_id
JOIN users u ON u.id = bs.user_id
WHERE bs.user_destination_id IN (SELECT id FROM user_destinations WHERE destination_id = 1)
ORDER BY bs.date DESC, bs.started_at DESC;

-- 2. List sessions for user numeric_id=100 (all destinations)
SELECT
    bs.id AS session_id,
    u.id AS user_id,
    u.numeric_id,
    u.name AS user_name,
    d.id AS destination_id,
    d.name AS destination_name,
    bs.date,
    bs.duration_minutes,
    bs.started_at,
    bs.ended_at
FROM broadcast_sessions bs
JOIN user_destinations ud ON ud.id = bs.user_destination_id
JOIN destinations d ON d.id = ud.destination_id
JOIN users u ON u.id = bs.user_id
WHERE u.numeric_id = '100'
ORDER BY bs.date DESC, bs.started_at DESC;

-- 3. List sessions for user numeric_id=100, destination id=1, with date range
SELECT
    bs.id AS session_id,
    u.id AS user_id,
    u.numeric_id,
    u.name AS user_name,
    d.id AS destination_id,
    d.name AS destination_name,
    bs.date,
    bs.duration_minutes,
    bs.started_at,
    bs.ended_at
FROM broadcast_sessions bs
JOIN user_destinations ud ON ud.id = bs.user_destination_id
JOIN destinations d ON d.id = ud.destination_id
JOIN users u ON u.id = bs.user_id
WHERE u.numeric_id = '100'
  AND bs.user_destination_id IN (SELECT id FROM user_destinations WHERE destination_id = 1)
  AND bs.date >= '2025-01-01'
  AND bs.date <= '2025-06-30'
ORDER BY bs.date DESC, bs.started_at DESC;