-- rollback for 000001_create_users.up.sql
DROP INDEX IF EXISTS users_email_idx;
DROP TABLE IF EXISTS users;
