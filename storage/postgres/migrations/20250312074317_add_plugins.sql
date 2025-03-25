-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS plugins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    metadata JSONB NOT NULL,
    server_endpoint VARCHAR(255) NOT NULL,
    pricing_id UUID NOT NULL,
    CONSTRAINT fk_pricing FOREIGN KEY (pricing_id) REFERENCES pricings(id) ON DELETE SET NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS plugins;
-- +goose StatementEnd
