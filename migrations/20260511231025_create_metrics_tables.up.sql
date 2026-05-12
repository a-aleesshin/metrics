CREATE TABLE IF NOT EXISTS metric(
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(70) NOT NULL,
    gauge_value DOUBLE PRECISION,
    counter_value BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    unique(name, type),

    CONSTRAINT check_value_filled CHECK (
        (type = 'gauge' AND gauge_value IS NOT NULL AND counter_value IS NULL) OR
        (type = 'counter' AND counter_value IS NOT NULL AND gauge_value IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_metric_name ON metric(name);
CREATE INDEX IF NOT EXISTS idx_metric_type ON metric(type);