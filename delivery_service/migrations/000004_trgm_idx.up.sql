DROP INDEX IF EXISTS products_tsv_idx;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX trgm_idx ON products USING GIN (description gin_trgm_ops);

'goo' -> '  g', ' go'
