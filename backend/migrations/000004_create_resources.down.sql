-- DB-016 rollback; DESTRUCTIVE: deletes all real resources and their data.
-- Do not CASCADE: fail safely if later migrations add dependent tables.
DROP TABLE public.resources;
