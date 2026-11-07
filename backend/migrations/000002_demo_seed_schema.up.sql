-- DB-011: isolated development fixture schema.
-- This is NOT the canonical application schema; production tables follow later.
CREATE SCHEMA demo_seed;

CREATE TABLE demo_seed.users (
    id INTEGER PRIMARY KEY,
    full_name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL CHECK (role IN ('admin', 'manager', 'employee'))
);

CREATE TABLE demo_seed.desks (
    id INTEGER PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    floor INTEGER NOT NULL,
    zone TEXT NOT NULL,
    has_monitor BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE demo_seed.meeting_rooms (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    floor INTEGER NOT NULL,
    capacity INTEGER NOT NULL CHECK (capacity > 0)
);

CREATE TABLE demo_seed.parking_spaces (
    id INTEGER PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    covered BOOLEAN NOT NULL DEFAULT FALSE,
    electric_charging BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE demo_seed.equipment (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    category TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity >= 0)
);
