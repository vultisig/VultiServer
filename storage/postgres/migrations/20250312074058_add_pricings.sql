-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'pricing_type') THEN
        CREATE TYPE pricing_type AS ENUM ('FREE', 'SINGLE', 'RECURRING', 'PER_TX');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'pricing_frequency') THEN
        CREATE TYPE pricing_frequency AS ENUM ('ANNUAL', 'MONTHLY', 'WEEKLY');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'pricing_metric') THEN
        CREATE TYPE pricing_metric AS ENUM ('FIXED', 'PERCENTAGE');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS pricings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type pricing_type NOT NULL,
    frequency pricing_frequency DEFAULT NULL,
    amount DOUBLE PRECISION NOT NULL,
    metric pricing_metric NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pricings;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'pricing_type') THEN
        DROP TYPE pricing_type;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'pricing_frequency') THEN
        DROP TYPE pricing_frequency;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'pricing_metric') THEN
        DROP TYPE pricing_metric;
    END IF;
END $$;
-- +goose StatementEnd
