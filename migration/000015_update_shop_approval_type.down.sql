ALTER TABLE shops DROP COLUMN approval_status;
DROP TYPE approval_status;
ALTER TABLE shops ADD COLUMN approved BOOLEAN NOT NULL DEFAULT false;