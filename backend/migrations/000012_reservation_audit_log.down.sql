-- Downgrade removes audit history; reservation history remains intact.
DROP TABLE public.reservation_audit_log;
