-- +goose Up
CREATE TABLE IF NOT EXISTS homepage(
    id BIGSERIAL PRIMARY KEY,
    title text NOT NULL,
    subtitle text NOT NULL,
    home_image text NOT NULL,
    home_image_alt text NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS homepage;
