CREATE TABLE exchange_rates (
    id SERIAL PRIMARY KEY,
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    rate NUMERIC(18,6) NOT NULL,
    source VARCHAR(50) DEFAULT 'manual',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(from_currency, to_currency)
);

INSERT INTO exchange_rates (from_currency, to_currency, rate, source) VALUES
    ('RUB', 'USD', 0.011000, 'initial'),
    ('USD', 'RUB', 90.910000, 'initial'),
    ('RUB', 'EUR', 0.010000, 'initial'),
    ('EUR', 'RUB', 100.000000, 'initial'),
    ('USD', 'EUR', 0.920000, 'initial'),
    ('EUR', 'USD', 1.090000, 'initial');

ALTER TABLE transactions ADD COLUMN IF NOT EXISTS converted_amount NUMERIC(18,2);
