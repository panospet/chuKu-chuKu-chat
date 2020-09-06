CREATE TABLE users(
   id SERIAL PRIMARY KEY,
   name VARCHAR(512) UNIQUE NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_name ON users (name);