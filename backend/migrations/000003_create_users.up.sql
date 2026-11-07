-- DB-015: canonical application users table, separate from demo_seed.users.
-- PostgreSQL 17 includes gen_random_uuid() without additional extensions.
CREATE TABLE public.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'employee',
    department TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_email_format_check
        CHECK (email = BTRIM(email) AND LENGTH(email) > 3 AND email LIKE '%@%'),
    CONSTRAINT users_password_hash_nonempty_check
        CHECK (LENGTH(BTRIM(password_hash)) > 0),
    CONSTRAINT users_first_name_nonempty_check
        CHECK (LENGTH(BTRIM(first_name)) > 0),
    CONSTRAINT users_last_name_nonempty_check
        CHECK (LENGTH(BTRIM(last_name)) > 0),
    CONSTRAINT users_role_check
        CHECK (role IN ('admin', 'manager', 'employee')),
    CONSTRAINT users_department_nonempty_check
        CHECK (department IS NULL OR LENGTH(BTRIM(department)) > 0)
);

-- Case-insensitive email uniqueness without an extra PostgreSQL extension.
CREATE UNIQUE INDEX users_email_lower_unique_idx
    ON public.users (LOWER(email));
