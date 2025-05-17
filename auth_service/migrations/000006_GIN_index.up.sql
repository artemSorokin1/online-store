DROP INDEX IF EXISTS username_index;

ALTER TABLE users ADD COLUMN IF NOT EXISTS username_tsv tsvector;

UPDATE users SET username_tsv = to_tsvector('simple', username);

CREATE INDEX idx_users_username_tsv ON users USING GIN(username_tsv);

