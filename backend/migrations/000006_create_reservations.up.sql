-- DB-018: canonical reservations connecting application users to bookable resources.
-- DB-019 will add advanced indexes and cross-reservation conflict rules.
CREATE TABLE public.reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users (id),
    resource_id UUID NOT NULL REFERENCES public.resources (id),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',
    check_in_code TEXT,
    confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT reservations_time_order_check
        CHECK (ends_at > starts_at),
    CONSTRAINT reservations_status_check
        CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'COMPLETED')),
    CONSTRAINT reservations_check_in_code_nonempty_check
        CHECK (check_in_code IS NULL OR (
            LENGTH(BTRIM(check_in_code)) > 0 AND check_in_code = BTRIM(check_in_code)
        ))
);
