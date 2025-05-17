DROP INDEX IF EXISTS idx_users_username_tsv;

ALTER TABLE users DROP column IF EXISTS username_tsv;