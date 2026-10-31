-- +goose Up
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username text NOT NULL,
    email text UNIQUE NOT NULL,
    password_hash text NOT NULL
);

-- +goose Down
DROP TABLE users;
