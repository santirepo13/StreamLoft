-- List all users
-- Usage: run the query below to view all users with destination count

-- Example:
SELECT
    u.id,
    u.numeric_id,
    u.name,
    u.stream_key,
    u.created_at,
    u.updated_at,
    COUNT(ud.id) AS destination_count
FROM users u
LEFT JOIN user_destinations ud ON ud.user_id = u.id
GROUP BY u.id
ORDER BY u.numeric_id::integer;