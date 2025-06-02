CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fullname TEXT NOT NULL DEFAULT '',
    username VARCHAR(100) UNIQUE NOT NULL,
    passhash TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone VARCHAR(50) NOT NULL DEFAULT '',
    city VARCHAR(100) NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    role VARCHAR(10) NOT NULL CHECK (role IN ('customer', 'seller')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS user_seller_username_hash_idx ON users USING HASH (username) WHERE role = 'seller';
