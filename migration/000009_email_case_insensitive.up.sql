-- Up Migration
CREATE EXTENSION IF NOT EXISTS citext;

ALTER TABLE users
    ALTER COLUMN email TYPE CITEXT,
    DROP CONSTRAINT users_email_key, -- Drop the old UNIQUE constraint
    ADD CONSTRAINT users_email_key UNIQUE (email);

