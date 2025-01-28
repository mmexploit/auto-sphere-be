CREATE TYPE role AS ENUM ('ADMIN', 'OPERATOR', 'SALES');

CREATE TABLE IF NOT EXISTS users {
    id SERIAL PRIMARY KEY,
    name text NOT NULL,
    email UNIQUE text,
    phone_number UNIQUE text NOT NULL,
    role role NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
}