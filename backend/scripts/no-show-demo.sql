-- Run only in the local development database. Creates a dedicated resource
-- and an already-overdue test reservation for an existing signed-in user.
-- Example from repo root:
-- docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U deskbooker -d deskbooker -v user_id=UUID < backend/scripts/no-show-demo.sql
\if :{?user_id}
\else
\echo 'Supply -v user_id=UUID for an existing local test user.'
\quit 1
\endif
BEGIN;
INSERT INTO public.resources(name,type,floor,capacity)
VALUES('No-show demo ' || gen_random_uuid()::text,'DESK',0,1)
RETURNING id AS demo_resource_id \gset
-- This deliberately bypasses future-only API validation to test expiration
-- without waiting 15 minutes. No other reservation or user is modified.
INSERT INTO public.reservations(user_id,resource_id,starts_at,ends_at,status,confirmed_at)
VALUES(:'user_id'::uuid,:'demo_resource_id'::uuid,
       statement_timestamp()-interval '16 minutes',
       statement_timestamp()+interval '1 hour','CONFIRMED',statement_timestamp())
RETURNING id AS demo_reservation_id \gset
COMMIT;
SELECT :'demo_resource_id' AS resource_id, :'demo_reservation_id' AS reservation_id;
