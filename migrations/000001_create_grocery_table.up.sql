-- Filename : 000001_create_grocery_table.up.sql

CREATE TABLE IF NOT EXISTS grocery (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    item text NOT NULL,
    location text NOT NULL, 
    price text NOT NULL, 
    address text NOT NULL, 
    phone text NOT NULL, 
    contact text NOT NULL, 
    email text NOT NULL, 
    website text NOT NULL,
    version int NOT NULL DEFAULT 1
);