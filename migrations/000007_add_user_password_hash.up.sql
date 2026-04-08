ALTER TABLE users
    ADD COLUMN IF NOT EXISTS password_hash TEXT;

UPDATE users
SET password_hash = '$2a$10$.y6NBqt4A8h.eaaIeLrJ2uu40vk3rh6Rz6/emwrrKMv0aS2wnrvcy'
WHERE password_hash IS NULL;

ALTER TABLE users
    ALTER COLUMN password_hash SET NOT NULL;
