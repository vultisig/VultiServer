-- +goose Up
-- +goose StatementBegin
CREATE TYPE transaction_status AS ENUM (
    'PENDING',
    'SIGNING_FAILED',
    'SIGNED',
    'BROADCAST',
    'MINED',
    'REJECTED'
);
CREATE TABLE transaction_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL REFERENCES plugin_policies(id),
    tx_body TEXT NOT NULL,
    tx_hash TEXT NOT NULL,
    status transaction_status NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB,
    error_message TEXT,
    CONSTRAINT fk_policy FOREIGN KEY (policy_id) REFERENCES plugin_policies(id)
);
CREATE INDEX idx_transaction_history_policy_id ON transaction_history(policy_id);
CREATE INDEX idx_transaction_history_status ON transaction_history(status);
CREATE INDEX idx_transaction_history_tx_hash ON transaction_history(tx_hash);
ALTER TABLE transaction_history
    ADD CONSTRAINT unique_tx_hash UNIQUE (tx_hash);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transaction_history;
DROP TYPE IF EXISTS transaction_status;
-- +goose StatementEnd