-- +goose Up
-- +goose StatementBegin
-- Create enum type for plugin_type
CREATE TYPE plugin_type AS ENUM ('payroll');

CREATE TABLE plugin_policies (
    id UUID PRIMARY KEY,
    public_key TEXT NOT NULL,
    plugin_id TEXT NOT NULL,
    plugin_version TEXT NOT NULL,
    policy_version TEXT NOT NULL,
    plugin_type plugin_type NOT NULL,
    signature TEXT NOT NULL,
    policy JSONB NOT NULL
);

-- Index for faster lookups on plugin_id
CREATE INDEX idx_plugin_policies_plugin_id ON plugin_policies(plugin_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS plugin_policies;

DROP TYPE IF EXISTS plugin_type;
-- +goose StatementEnd
