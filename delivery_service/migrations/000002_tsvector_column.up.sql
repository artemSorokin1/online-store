ALTER TABLE products ADD COLUMN IF NOT EXISTS description_tsv tsvector;

UPDATE products SET description_tsv = to_tsvector('english', description);
