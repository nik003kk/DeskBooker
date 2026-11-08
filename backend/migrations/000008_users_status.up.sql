-- DB-029: explicit account status for the authenticated profile.
ALTER TABLE public.users
    ADD COLUMN status TEXT NOT NULL DEFAULT 'active',
    ADD CONSTRAINT users_status_check CHECK (status IN ('active', 'inactive', 'suspended'));
