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

-- Create time_triggers table for scheduling
CREATE TABLE time_triggers (
    id SERIAL PRIMARY KEY,
    policy_id UUID NOT NULL REFERENCES plugin_policies(id),
    cron_expression TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    frequency TEXT NOT NULL,
    last_execution TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for time_triggers
CREATE INDEX idx_time_triggers_policy_id ON time_triggers(policy_id);
CREATE INDEX idx_time_triggers_start_time ON time_triggers(start_time);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS time_triggers;
DROP TABLE IF EXISTS plugin_policies;
DROP TYPE IF EXISTS plugin_type;
-- +goose StatementEnd