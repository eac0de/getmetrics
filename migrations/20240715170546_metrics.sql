-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    metrics (
        id TEXT,
        type TEXT NOT NULL,
        delta BIGINT,
        value DOUBLE PRECISION
    );

CREATE UNIQUE INDEX metrics_unique_idx ON metrics (id, type);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE metrics;

-- +goose StatementEnd