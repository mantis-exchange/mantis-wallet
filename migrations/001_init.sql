CREATE TABLE IF NOT EXISTS addresses (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    chain VARCHAR(10) NOT NULL,
    address VARCHAR(255) NOT NULL,
    type VARCHAR(10) NOT NULL DEFAULT 'DEPOSIT',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, chain, type)
);

CREATE TABLE IF NOT EXISTS deposits (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    chain VARCHAR(10) NOT NULL,
    asset VARCHAR(20) NOT NULL,
    tx_hash VARCHAR(255) UNIQUE NOT NULL,
    amount VARCHAR(40) NOT NULL,
    confirmations INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deposits_user ON deposits(user_id);
CREATE INDEX IF NOT EXISTS idx_deposits_status ON deposits(status);

CREATE TABLE IF NOT EXISTS withdrawals (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    chain VARCHAR(10) NOT NULL,
    asset VARCHAR(20) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    amount VARCHAR(40) NOT NULL,
    fee VARCHAR(40) NOT NULL DEFAULT '0',
    tx_hash VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_withdrawals_user ON withdrawals(user_id);
CREATE INDEX IF NOT EXISTS idx_withdrawals_status ON withdrawals(status);
