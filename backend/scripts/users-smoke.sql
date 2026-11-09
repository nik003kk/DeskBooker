-- DB-015 integration smoke tests, executed only against a migrated development DB.
-- All inserts are rolled back. psql must run with ON_ERROR_STOP=1.
BEGIN;

DO $test$
DECLARE
    sample_id uuid;
    duplicate_rejected boolean := false;
    invalid_role_rejected boolean := false;
    empty_hash_rejected boolean := false;
    invalid_email_rejected boolean := false;
BEGIN
    INSERT INTO public.users (email, password_hash, first_name, last_name, role, department)
    VALUES ('db015-smoke@example.test', 'example-hash-not-a-real-password', 'Test', 'User', 'employee', 'QA')
    RETURNING id INTO sample_id;

    IF sample_id IS NULL OR (
      SELECT COUNT(*) FROM public.users WHERE id = sample_id AND created_at IS NOT NULL
    ) <> 1 THEN
        RAISE EXCEPTION 'valid user was not inserted correctly';
    END IF;

    -- Unique index must reject differently capitalized duplicate email.
    BEGIN
        INSERT INTO public.users (email, password_hash, first_name, last_name)
        VALUES ('DB015-SMOKE@example.test', 'example-hash', 'Other', 'User');
    EXCEPTION WHEN unique_violation THEN
        duplicate_rejected := true;
    END;
    IF NOT duplicate_rejected THEN
        RAISE EXCEPTION 'case-insensitive duplicate email was accepted';
    END IF;

    BEGIN
        INSERT INTO public.users (email, password_hash, first_name, last_name, role)
        VALUES ('db015-bad-role@example.test', 'example-hash', 'Bad', 'Role', 'superuser');
    EXCEPTION WHEN check_violation THEN
        invalid_role_rejected := true;
    END;
    IF NOT invalid_role_rejected THEN
        RAISE EXCEPTION 'invalid role was accepted';
    END IF;

    BEGIN
        INSERT INTO public.users (email, password_hash, first_name, last_name)
        VALUES ('db015-no-hash@example.test', ' ', 'No', 'Hash');
    EXCEPTION WHEN check_violation THEN
        empty_hash_rejected := true;
    END;
    IF NOT empty_hash_rejected THEN
        RAISE EXCEPTION 'empty password_hash was accepted';
    END IF;

    BEGIN
        INSERT INTO public.users (email, password_hash, first_name, last_name)
        VALUES ('db015-invalid-email', 'example-hash', 'Bad', 'Email');
    EXCEPTION WHEN check_violation THEN
        invalid_email_rejected := true;
    END;
    IF NOT invalid_email_rejected THEN
        RAISE EXCEPTION 'invalid email was accepted';
    END IF;

    RAISE NOTICE 'DB-015 smoke tests passed: UUID/defaults, duplicate email, role, hash, email';
END
$test$;

ROLLBACK;
