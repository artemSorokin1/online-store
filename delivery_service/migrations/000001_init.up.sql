CREATE TABLE IF NOT EXISTS products (
      ID SERIAL PRIMARY KEY,
      Name VARCHAR(255) NOT NULL,
      Price INT NOT NULL,
      Sizes INTEGER[] NOT NULL,
      ImageURL VARCHAR(255),
      Description TEXT
);

CREATE TABLE IF NOT EXISTS orders (
      ID SERIAL PRIMARY KEY,
      ProductIDs INTEGER[] NOT NULL,
      Sizes INTEGER[] NOT NULL,
      user_id INT NOT NULL,
      Total INT NOT NULL,
      Name VARCHAR(255) NOT NULL,
      Address VARCHAR(255) NOT NULL,
      UserPhone VARCHAR(255) NOT NULL,
      OrderDate TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS cart (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    total INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS cart_items (
    id SERIAL PRIMARY KEY,
    cart_id INT NOT NULL,
    product_id INT NOT NULL,
    size INT NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    FOREIGN KEY (cart_id) REFERENCES cart(id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    UNIQUE (product_id, cart_id)
);

CREATE TABLE IF NOT EXISTS user_actions (
    order_id INT NOT NULL,
    user_id INT NOT NULL,
    action VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id)
);


