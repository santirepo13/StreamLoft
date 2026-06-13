-- Edit a platform destination name or RTMP URL
-- Usage: run the query below, edit the SET and WHERE values, then execute

-- Examples:
-- 1. Update YouTube RTMP URL
UPDATE destinations SET name = 'YouTube', rtmp_url = 'rtmp://a.rtmp.youtube.com/live2' WHERE id = 1;

-- 2. Update Twitch RTMP URL
UPDATE destinations SET name = 'Twitch', rtmp_url = 'rtmp://live.twitch.tv/app' WHERE id = 2;