-- Create a global platform destination (YouTube, Twitch, Facebook, etc.)
-- Run this once per platform, then assign to users with assign_destination.sql
-- Usage: run the query below, edit the VALUES, then execute

-- Examples:
-- 1. Create YouTube destination
INSERT INTO destinations (name, rtmp_url, created_at) VALUES ('YouTube', 'rtmp://a.rtmp.youtube.com/live2', NOW());

-- 2. Create Twitch destination
INSERT INTO destinations (name, rtmp_url, created_at) VALUES ('Twitch', 'rtmp://live.twitch.tv/app', NOW());

-- 3. Create Facebook destination
INSERT INTO destinations (name, rtmp_url, created_at) VALUES ('Facebook', 'rtmps://live-api-s.facebook.com:443/rtmp', NOW());