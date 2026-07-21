CREATE TABLE IF NOT EXISTS rate_history (
    id BIGSERIAL PRIMARY KEY,

    currency CHAR(3) NOT NULL,
    value NUMERIC(18, 6) NOT NULL CHECK (value > 0),

    effective_at TIMESTAMPTZ NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT rate_history_currency_check
        CHECK (currency IN ('USD', 'EUR', 'CNY')),

    CONSTRAINT rate_history_currency_effective_unique
        UNIQUE (currency, effective_at)
);

CREATE INDEX IF NOT EXISTS idx_rate_history_currency_effective_at
ON rate_history(currency, effective_at DESC);
