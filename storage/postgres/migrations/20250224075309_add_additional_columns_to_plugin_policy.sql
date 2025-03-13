-- +goose Up
-- +goose StatementBegin
ALTER TABLE plugin_policies
ADD COLUMN is_ecdsa BOOLEAN DEFAULT TRUE;
ALTER TABLE plugin_policies
ADD COLUMN chain_code_hex TEXT NOT NULL;
ALTER TABLE plugin_policies
ADD COLUMN derive_path TEXT NOT NULL;
ALTER TABLE plugin_policies
ADD COLUMN active BOOLEAN DEFAULT TRUE;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE plugin_policies DROP COLUMN active,
  DROP COLUMN derive_path,
  DROP COLUMN chain_code_hex,
  DROP COLUMN is_ecdsa;
-- +goose StatementEnd