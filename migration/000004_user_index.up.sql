ALTER TABLE users ADD COLUMN name_tsvector tsvector 
GENERATED ALWAYS AS (to_tsvector('simple', name)) STORED;

CREATE INDEX IF NOT EXISTS user_name_idx ON users USING GIN(name_tsvector);
CREATE INDEX IF NOT EXISTS user_role_idx ON users USING BTREE(role);
