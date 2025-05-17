INSERT INTO users (username, email, passhash)
SELECT
    'user' || ' ' || gs,
    'user_' || gs || '@example.com',
    'static_hash'
FROM generate_series(1, 5000000) AS gs;