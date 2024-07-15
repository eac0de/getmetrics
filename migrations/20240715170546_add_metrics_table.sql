-- +goose Up
-- +goose StatementBegin
CREATE TABLE metrics (
    id TEXT,
    m_type TEXT NOT NULL,
    delta INT,
    value DOUBLE PRECISION
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE metrics;
-- +goose StatementEnd
