-- DB-017: destructive rollback for associated features only.
-- The parent public.resources table is deliberately preserved.
DROP TABLE public.resource_features;
