-- DB-016: integration check against a local migrated PostgreSQL database.
-- Entire test is rolled back; run psql with -v ON_ERROR_STOP=1.
BEGIN;

DO $test$
DECLARE
    desk_id uuid;
    invalid_type_rejected boolean := false;
    invalid_capacity_rejected boolean := false;
    empty_name_rejected boolean := false;
    duplicate_name_rejected boolean := false;
    missing_name_rejected boolean := false;
BEGIN
    -- Test defaults, optional floor and all supported resource types.
    INSERT INTO public.resources (name, type, floor)
    VALUES ('DB016 Demo Desk', 'DESK', 1)
    RETURNING id INTO desk_id;

    IF desk_id IS NULL OR NOT EXISTS (
        SELECT 1
        FROM public.resources
        WHERE id = desk_id
          AND capacity = 1
          AND is_active = TRUE
          AND created_at IS NOT NULL
          AND floor = 1
    ) THEN
        RAISE EXCEPTION 'resource defaults or UUID generation failed';
    END IF;

    INSERT INTO public.resources (name, type, floor, capacity)
    VALUES ('DB016 Demo Meeting Room', 'MEETING_ROOM', 2, 8);

    INSERT INTO public.resources (name, type, floor, capacity, is_active)
    VALUES ('DB016 Demo Parking', 'PARKING_SPOT', NULL, 1, FALSE);

    IF (
        SELECT COUNT(*) FROM public.resources WHERE name LIKE 'DB016 Demo %'
    ) <> 3 THEN
        RAISE EXCEPTION 'expected three valid resource types';
    END IF;

    BEGIN
        INSERT INTO public.resources (name, type)
        VALUES ('DB016 Bad Type', 'EQUIPMENT');
    EXCEPTION WHEN check_violation THEN
        invalid_type_rejected := TRUE;
    END;
    IF NOT invalid_type_rejected THEN
        RAISE EXCEPTION 'unsupported type was accepted';
    END IF;

    BEGIN
        INSERT INTO public.resources (name, type, capacity)
        VALUES ('DB016 No Capacity', 'MEETING_ROOM', 0);
    EXCEPTION WHEN check_violation THEN
        invalid_capacity_rejected := TRUE;
    END;
    IF NOT invalid_capacity_rejected THEN
        RAISE EXCEPTION 'zero capacity was accepted';
    END IF;

    BEGIN
        INSERT INTO public.resources (name, type)
        VALUES ('  ', 'DESK');
    EXCEPTION WHEN check_violation THEN
        empty_name_rejected := TRUE;
    END;
    IF NOT empty_name_rejected THEN
        RAISE EXCEPTION 'blank resource name was accepted';
    END IF;

    BEGIN
        INSERT INTO public.resources (name, type)
        VALUES ('db016 demo desk', 'DESK');
    EXCEPTION WHEN unique_violation THEN
        duplicate_name_rejected := TRUE;
    END;
    IF NOT duplicate_name_rejected THEN
        RAISE EXCEPTION 'case-insensitive duplicate name was accepted';
    END IF;

    BEGIN
        INSERT INTO public.resources (type)
        VALUES ('DESK');
    EXCEPTION WHEN not_null_violation THEN
        missing_name_rejected := TRUE;
    END;
    IF NOT missing_name_rejected THEN
        RAISE EXCEPTION 'missing resource name was accepted';
    END IF;

    RAISE NOTICE 'DB-016 smoke tests passed: UUID, defaults, 3 types, constraints, duplicate name';
END
$test$;

ROLLBACK;
