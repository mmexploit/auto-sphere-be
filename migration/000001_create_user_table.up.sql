CREATE TYPE role AS ENUM ('ADMIN', 'OPERATOR', 'SALES');

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE, -- Moved UNIQUE constraint after type definition
    phone_number TEXT UNIQUE NOT NULL,
    role role NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
