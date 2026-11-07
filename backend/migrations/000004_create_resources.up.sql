-- DB-016: Canonical bookable resources. The demo_seed schema is not affected.
CREATE TABLE public.resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    floor INTEGER,
    capacity INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT resources_name_nonempty_check
        CHECK (LENGTH(BTRIM(name)) > 0 AND name = BTRIM(name)),
    CONSTRAINT resources_type_check
        CHECK (type IN ('DESK', 'MEETING_ROOM', 'PARKING_SPOT')),
    CONSTRAINT resources_capacity_positive_check
        CHECK (capacity > 0)
);

-- A resource's name is unique regardless of letter case.
CREATE UNIQUE INDEX resources_name_lower_unique_idx
    ON public.resources (LOWER(name));

-- Supports filtering available resources by type and optionally floor.
CREATE INDEX resources_active_type_floor_idx
    ON public.resources (type, floor) WHERE is_active = TRUE;
