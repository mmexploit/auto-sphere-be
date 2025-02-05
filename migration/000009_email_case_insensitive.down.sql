-- Down Migration (Revert Changes)
ALTER TABLE users
    ALTER COLUMN email TYPE TEXT,
    DROP CONSTRAINT users_email_key,
    ADD CONSTRAINT users_email_key UNIQUE (email);