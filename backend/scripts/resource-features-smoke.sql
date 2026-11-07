-- DB-017 live PostgreSQL test. Execute with psql -X -v ON_ERROR_STOP=1.
-- Everything happens inside a transaction and is rolled back.
BEGIN;

DO $smoke$
DECLARE
    desk_id uuid;
    second_desk_id uuid;
    missing_resource_rejected boolean := false;
    duplicate_feature_rejected boolean := false;
    blank_feature_rejected boolean := false;
    missing_feature_rejected boolean := false;
BEGIN
    INSERT INTO public.resources (name, type, floor)
    VALUES ('DB017 Smoke Desk A', 'DESK', 1) RETURNING id INTO desk_id;

    INSERT INTO public.resources (name, type, floor)
    VALUES ('DB017 Smoke Desk B', 'DESK', 1) RETURNING id INTO second_desk_id;

    -- A resource can have multiple features and different resources can share one.
    INSERT INTO public.resource_features (resource_id, feature)
    VALUES (desk_id, 'monitor'), (desk_id, 'docking_station'),
           (second_desk_id, 'monitor');

    IF (SELECT COUNT(*) FROM public.resource_features WHERE resource_id = desk_id) <> 2
       OR (SELECT COUNT(*) FROM public.resource_features WHERE feature = 'monitor') <> 2 THEN
        RAISE EXCEPTION 'multiple/shared features failed';
    END IF;

    -- Duplicate association must fail even with a different letter case.
    BEGIN
        INSERT INTO public.resource_features (resource_id, feature)
        VALUES (desk_id, 'MONITOR');
    EXCEPTION WHEN unique_violation THEN
        duplicate_feature_rejected := TRUE;
    END;
    IF NOT duplicate_feature_rejected THEN
        RAISE EXCEPTION 'duplicate feature was accepted';
    END IF;

    BEGIN
        INSERT INTO public.resource_features (resource_id, feature)
        VALUES (desk_id, '   ');
    EXCEPTION WHEN check_violation THEN
        blank_feature_rejected := TRUE;
    END;
    IF NOT blank_feature_rejected THEN
        RAISE EXCEPTION 'blank feature was accepted';
    END IF;

    BEGIN
        INSERT INTO public.resource_features (resource_id, feature)
        VALUES (desk_id, NULL);
    EXCEPTION WHEN not_null_violation THEN
        missing_feature_rejected := TRUE;
    END;
    IF NOT missing_feature_rejected THEN
        RAISE EXCEPTION 'NULL feature was accepted';
    END IF;

    BEGIN
        INSERT INTO public.resource_features (resource_id, feature)
        VALUES ('00000000-0000-0000-0000-000000000000'::uuid, 'projector');
    EXCEPTION WHEN foreign_key_violation THEN
        missing_resource_rejected := TRUE;
    END;
    IF NOT missing_resource_rejected THEN
        RAISE EXCEPTION 'orphan feature was accepted';
    END IF;

    -- Cascade deletion affects features of the deleted resource, not other resources.
    DELETE FROM public.resources WHERE id = desk_id;
    IF EXISTS (SELECT 1 FROM public.resource_features WHERE resource_id = desk_id) THEN
        RAISE EXCEPTION 'features were not removed with their parent resource';
    END IF;
    IF (SELECT COUNT(*) FROM public.resource_features WHERE resource_id = second_desk_id) <> 1 THEN
        RAISE EXCEPTION 'unrelated features were removed';
    END IF;

    RAISE NOTICE 'DB-017 smoke tests passed: multiple/shared features, duplicate, validation, FK, cascade';
END
$smoke$;

ROLLBACK;
