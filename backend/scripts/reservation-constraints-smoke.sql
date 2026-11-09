-- DB-019 PostgreSQL integration test, rollback-only.
-- Run with psql -X -v ON_ERROR_STOP=1 against migration version 7.
BEGIN;

DO $smoke$
DECLARE
    test_user uuid;
    first_resource uuid;
    second_resource uuid;
    first_reservation uuid;
    cancelled_reservation uuid;
    overlap_rejected boolean := FALSE;
    update_overlap_rejected boolean := FALSE;
    code_rejected boolean := FALSE;
    other_user_missing_rejected boolean := FALSE;
    other_resource_missing_rejected boolean := FALSE;
BEGIN
    INSERT INTO public.users (email, password_hash, first_name, last_name)
    VALUES ('db019-smoke@example.test', 'test-hash-only', 'Smoke', 'Tests')
    RETURNING id INTO test_user;

    INSERT INTO public.resources (name, type)
    VALUES ('DB019 Room A', 'MEETING_ROOM')
    RETURNING id INTO first_resource;

    INSERT INTO public.resources (name, type)
    VALUES ('DB019 Room B', 'MEETING_ROOM')
    RETURNING id INTO second_resource;

    INSERT INTO public.reservations
        (user_id, resource_id, starts_at, ends_at, status, check_in_code)
    VALUES
        (test_user, first_resource, '2031-03-01 09:00+00', '2031-03-01 10:00+00',
         'CONFIRMED', 'DB019-CODE-1')
    RETURNING id INTO first_reservation;

    -- Same resource and intersecting active intervals must be rejected.
    BEGIN
        INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at)
        VALUES (test_user, first_resource, '2031-03-01 09:30+00', '2031-03-01 10:30+00');
    EXCEPTION WHEN exclusion_violation THEN
        overlap_rejected := TRUE;
    END;
    IF NOT overlap_rejected THEN
        RAISE EXCEPTION 'overlapping active reservation was accepted';
    END IF;

    -- Adjacent intervals [09,10) and [10,11) are valid.
    INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at, status)
    VALUES (test_user, first_resource, '2031-03-01 10:00+00', '2031-03-01 11:00+00', 'PENDING');

    -- Different resource may be booked at the same time.
    INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at)
    VALUES (test_user, second_resource, '2031-03-01 09:30+00', '2031-03-01 10:30+00');

    -- A cancelled reservation does not prevent a new booking of its interval.
    INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at, status)
    VALUES (test_user, first_resource, '2031-03-01 09:30+00', '2031-03-01 10:00+00', 'CANCELLED')
    RETURNING id INTO cancelled_reservation;

    -- Switching cancellation to active should be blocked too.
    BEGIN
        UPDATE public.reservations SET status = 'CONFIRMED'
        WHERE id = cancelled_reservation;
    EXCEPTION WHEN exclusion_violation THEN
        update_overlap_rejected := TRUE;
    END;
    IF NOT update_overlap_rejected THEN
        RAISE EXCEPTION 'activation of overlapping reservation was accepted';
    END IF;

    -- Completed bookings do not block later active bookings, even with identical times.
    INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at, status)
    VALUES (test_user, first_resource, '2031-03-01 09:00+00', '2031-03-01 10:00+00', 'COMPLETED');

    -- A check-in code is unique regardless of case and resource.
    BEGIN
        INSERT INTO public.reservations
            (user_id, resource_id, starts_at, ends_at, check_in_code)
        VALUES (test_user, second_resource, '2031-03-01 11:00+00', '2031-03-01 12:00+00',
                'db019-code-1');
    EXCEPTION WHEN unique_violation THEN
        code_rejected := TRUE;
    END;
    IF NOT code_rejected THEN
        RAISE EXCEPTION 'duplicate check-in code was accepted';
    END IF;

    -- Existing foreign keys from DB-018 must still work.
    BEGIN
        INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at)
        VALUES ('00000000-0000-0000-0000-000000000000'::uuid,
                first_resource, '2031-03-01 15:00+00', '2031-03-01 16:00+00');
    EXCEPTION WHEN foreign_key_violation THEN
        other_user_missing_rejected := TRUE;
    END;
    IF NOT other_user_missing_rejected THEN
        RAISE EXCEPTION 'unknown user was accepted';
    END IF;

    BEGIN
        INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at)
        VALUES (test_user, '00000000-0000-0000-0000-000000000000'::uuid,
                '2031-03-01 15:00+00', '2031-03-01 16:00+00');
    EXCEPTION WHEN foreign_key_violation THEN
        other_resource_missing_rejected := TRUE;
    END;
    IF NOT other_resource_missing_rejected THEN
        RAISE EXCEPTION 'unknown resource was accepted';
    END IF;

    IF (SELECT COUNT(*) FROM public.reservations WHERE id = first_reservation) <> 1 THEN
        RAISE EXCEPTION 'initial booking was lost';
    END IF;

    RAISE NOTICE 'DB-019 smoke tests passed: overlap, adjacency, resource isolation, statuses, code uniqueness, FKs';
END
$smoke$;

ROLLBACK;
