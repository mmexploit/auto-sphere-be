DROP INDEX IF EXISTS user_name_idx;
DROP INDEX IF EXISTS user_role_idx;
ALTER TABLE users DROP COLUMN IF EXISTS name_tsvector;
