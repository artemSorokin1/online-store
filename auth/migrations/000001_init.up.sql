CREATE TABLE customer (
    id UUID PRIMARY KEY,
    fullname VARCHAR(50),
    passhash VARCHAR(50),
    email VARCHAR(50),
    phone VARCHAR(50),
    city VARCHAR(50),
    addres VARCHAR(50),
    created_acc TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE seller (
    id UUID PRIMARY KEY,
    phone VARCHAR(50),
    email VARCHAR(50),
    fullname VARCHAR(50)
);
