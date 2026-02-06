-- 启用 uuid 扩展
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 创建 tokens 表
CREATE TABLE IF NOT EXISTS tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    address VARCHAR(42) UNIQUE NOT NULL,
    name VARCHAR(255),
    symbol VARCHAR(50),
    decimals INTEGER DEFAULT 18,
    pair_address VARCHAR(42),
    dex VARCHAR(50) DEFAULT 'pancakeswap',
    initial_liquidity DECIMAL(36, 18),
    risk_score INTEGER,
    risk_level VARCHAR(20),
    risk_details JSONB,
    is_honeypot BOOLEAN DEFAULT FALSE,
    buy_tax DECIMAL(5, 2),
    sell_tax DECIMAL(5, 2),
    creator_address VARCHAR(42),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建 trades 表
CREATE TABLE IF NOT EXISTS trades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_address VARCHAR(42) NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    token_symbol VARCHAR(50),
    type VARCHAR(10) NOT NULL,
    amount_in DECIMAL(36, 18),
    amount_out DECIMAL(36, 18),
    tx_hash VARCHAR(66) UNIQUE,
    status VARCHAR(20),
    gas_used DECIMAL(36, 18),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_tokens_created_at ON tokens(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tokens_risk_score ON tokens(risk_score);
CREATE INDEX IF NOT EXISTS idx_trades_user_address ON trades(user_address);
CREATE INDEX IF NOT EXISTS idx_trades_token_address ON trades(token_address);
