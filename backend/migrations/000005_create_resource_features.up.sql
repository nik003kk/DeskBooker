-- DB-017: each resource may have many named features or equipment attributes.
-- Free-text feature names avoid introducing a premature canonical equipment table.
CREATE TABLE public.resource_features (
    resource_id UUID NOT NULL
        REFERENCES public.resources (id) ON DELETE CASCADE,
    feature TEXT NOT NULL,
    CONSTRAINT resource_features_feature_nonempty_check
        CHECK (LENGTH(BTRIM(feature)) > 0 AND feature = BTRIM(feature))
);

-- A given resource cannot have the same feature twice (case insensitive).
-- Also makes lookups for the features of one resource efficient.
CREATE UNIQUE INDEX resource_features_resource_feature_unique_idx
    ON public.resource_features (resource_id, LOWER(feature));
