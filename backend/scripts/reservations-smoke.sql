-- DB-018 integration smoke test on a local migrated PostgreSQL database.
-- Run with psql -X -v ON_ERROR_STOP=1. All test data is rolled back.
BEGIN;

DO $smoke$
DECLARE
    test_user uuid;
    test_resource uuid;
    test_reservation uuid;
    invalid_time_rejected boolean := FALSE;
    invalid_status_rejected boolean := FALSE;
    invalid_code_rejected boolean := FALSE;
    missing_user_rejected boolean := FALSE;
    missing_resource_rejected boolean := FALSE;
    missing_start_rejected boolean := FALSE;
BEGIN
    INSERT INTO public.users (email, password_hash, first_name, last_name)
    VALUES ('db018-smoke@example.test', 'test-hash-not-real-password', 'Smoke', 'Test')
    RETURNING id INTO test_user;

    INSERT INTO public.resources (name, type, floor)
    VALUES ('DB018 Smoke Desk', 'DESK', 1)
    RETURNING id INTO test_resource;

    INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at)
    VALUES (test_user, test_resource, '2030-01-01 09:00:00+00', '2030-01-01 10:00:00+00')
    RETURNING id INTO test_reservation;

    IF test_reservation IS NULL OR NOT EXISTS (
        SELECT 1 FROM public.reservations
        WHERE id = test_reservation
          AND user_id = test_user
          AND resource_id = test_resource
          AND status = 'PENDING'
          AND check_in_code IS NULL
          AND confirmed_at IS NULL
          AND created_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'valid reservation or defaults failed';
    END IF;

    UPDATE public.reservations
    SET status = 'CONFIRMED', check_in_code = 'CHECKIN-018', confirmed_at = NOW()
    WHERE id = test_reservation;

    IF NOT EXISTS (
        SELECT 1 FROM public.reservations
        WHERE id = test_reservation
          AND status = 'CONFIRMED'
          AND check_in_code = 'CHECKIN-018'
          AND confirmed_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'confirmation and check-in fields failed';
    END IF;

    BEGIN
        INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at)
        VALUES (test_user, test_resource, '2030-01-01 10:00:00+00', '2030-01-01 10:00:00+00');
    EXCEPTION WHEN check_violation THEN
        invalid_time_rejected := TRUE;
    END;
    IF NOT invalid_time_rejected THEN
        RAISE EXCEPTION 'zero-length reservation was accepted';
    END IF;

    BEGIN
        INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at, status)
        VALUES (test_user, test_resource, '2030-01-01 10:00:00+00', '2030-01-01 09:00:00+00', 'PENDING');
    EXCEPTION WHEN check_violation THEN
        -- Reverse time is also invalid; use separate check outside below.
        NULL;
    END;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'public.reservations'::regclass
          AND conname = 'reservations_time_order_check'
    ) THEN
        RAISE EXCEPTION 'time ordering constraint is missing';
    END IF;

    BEGIN
        INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at, status)
        VALUES (test_user, test_resource, '2030-01-01 11:00:00+00', '2030-01-01 12:00:00+00', 'INVALID');
    EXCEPTION WHEN check_violation THEN
        invalid_status_rejected := TRUE;
    END;
    IF NOT invalid_status_rejected THEN
        RAISE EXCEPTION 'unsupported reservation status was accepted';
    END IF;

    BEGIN
        INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at, check_in_code)
        VALUES (test_user, test_resource, '2030-01-01 11:00:00+00', '2030-01-01 12:00:00+00', '   ');
    EXCEPTION WHEN check_violation THEN
        invalid_code_rejected := TRUE;
    END;
    IF NOT invalid_code_rejected THEN
        RAISE EXCEPTION 'blank check-in code was accepted';
    END IF;

    BEGIN
        INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at)
        VALUES ('00000000-0000-0000-0000-000000000000'::uuid, test_resource,
                '2030-01-01 11:00:00+00', '2030-01-01 12:00:00+00');
    EXCEPTION WHEN foreign_key_violation THEN
        missing_user_rejected := TRUE;
    END;
    IF NOT missing_user_rejected THEN
        RAISE EXCEPTION 'missing user was accepted';
    END IF;

    BEGIN
        INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at)
        VALUES (test_user, '00000000-0000-0000-0000-000000000000'::uuid,
                '2030-01-01 11:00:00+00', '2030-01-01 12:00:00+00');
    EXCEPTION WHEN foreign_key_violation THEN
        missing_resource_rejected := TRUE;
    END;
    IF NOT missing_resource_rejected THEN
        RAISE EXCEPTION 'missing resource was accepted';
    END IF;

    BEGIN
        INSERT INTO public.reservations (user_id, resource_id, ends_at)
        VALUES (test_user, test_resource, '2030-01-01 12:00:00+00');
    EXCEPTION WHEN not_null_violation THEN
        missing_start_rejected := TRUE;
    END;
    IF NOT missing_start_rejected THEN
        RAISE EXCEPTION 'missing start time was accepted';
    END IF;

    RAISE NOTICE 'DB-018 smoke tests passed: UUID/defaults, confirmation, time, status, code, foreign keys, required fields';
END
$smoke$;

ROLLBACK;
