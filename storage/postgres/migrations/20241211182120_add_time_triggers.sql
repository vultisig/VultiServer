-- +goose Up
-- +goose StatementBegin
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
-- +goose StatementEnd