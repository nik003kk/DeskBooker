CREATE INDEX reservations_analytics_starts_at_idx
    ON public.reservations (starts_at) INCLUDE (status, ends_at);
