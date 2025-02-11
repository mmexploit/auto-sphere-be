CREATE TYPE approval_status AS ENUM ('PENDING', 'APPROVED', 'DECLINED');
ALTER TABLE shops DROP COLUMN approved;
ALTER TABLE shops ADD COLUMN approval_status approval_status NOT NULL DEFAULT 'PENDING';