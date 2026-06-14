CREATE TYPE audit_action AS ENUM (
    'TRANSFER_CREATED',
    'TRANSFER_COMPLETED',
    'TRANSFER_FAILED',
    'TRANSFER_BLOCKED_BY_FRAUD',
    'ACCOUNT_CREATED',
    'ACCOUNT_BLOCKED',
    'ACCOUNT_UNBLOCKED',
    'USER_REGISTERED',
    'USER_VERIFIED'
);

CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action audit_action NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    old_value JSONB,
    new_value JSONB,
    ip_address VARCHAR(45),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX ix_audit_log_user_id ON audit_log(user_id);
CREATE INDEX ix_audit_log_created_at ON audit_log(created_at);
CREATE INDEX ix_audit_log_action ON audit_log(action);
CREATE INDEX ix_audit_log_entity ON audit_log(entity_type, entity_id);

CREATE TABLE fraud_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(id),
    verdict VARCHAR(20) NOT NULL,
    risk_score NUMERIC(5,2) NOT NULL,
    features_json JSONB,
    engine VARCHAR(50),
    checked_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX ix_fraud_checks_transaction_id ON fraud_checks(transaction_id);
