CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE shops (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    location TEXT NOT NULL,
    coordinate GEOGRAPHY(POINT, 4326) NOT NULL, -- Stores lat/lng with PostGIS
    category TEXT[] DEFAULT '{}',
    created_at TIMESTAMP DEFAULT now()
);

-- Index for fast spatial queries
CREATE INDEX shops_coordinate_idx ON shops USING GIST(coordinate);
