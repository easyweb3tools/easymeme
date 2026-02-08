export interface Token {
  id: string;
  address: string;
  name: string;
  symbol: string;
  decimals: number;
  pair_address: string;
  initial_liquidity: string;
  risk_score: number;
  risk_level: 'pending' | 'safe' | 'warning' | 'danger';
  is_honeypot: boolean;
  buy_tax: number;
  sell_tax: number;
  created_at: string;
}

export type GoldenDogToken = {
  address: string;
  name: string;
  symbol: string;
  pairAddress: string;
  liquidity: number;
  riskScore: number;
  riskLevel: 'pending' | 'safe' | 'warning' | 'danger';
  isGoldenDog: boolean;
  goldenDogScore: number;
  effectiveScore: number;
  timeDecayFactor: number;
  phase: 'EARLY' | 'PEAK' | 'DECLINING' | 'EXPIRED';
  createdAt: string;
  analyzedAt?: string | null;
};

export type TokenDetail = {
  id: string;
  address: string;
  name: string;
  symbol: string;
  pairAddress: string;
  dex: string;
  liquidity: number;
  creatorAddress: string;
  createdAt: string;
  analyzedAt?: string | null;
  riskScore: number;
  riskLevel: 'pending' | 'safe' | 'warning' | 'danger';
  isGoldenDog: boolean;
  goldenDogScore: number;
  effectiveScore: number;
  timeDecayFactor: number;
  phase: 'EARLY' | 'PEAK' | 'DECLINING' | 'EXPIRED';
  riskDetails?: Record<string, unknown>;
  analysisResult?: Record<string, unknown>;
};

export type AITrade = {
  id: string;
  token_address: string;
  token_symbol: string;
  type: 'BUY' | 'SELL';
  amount_in: string;
  amount_out: string;
  tx_hash: string;
  timestamp: string;
  status: string;
  gas_used: string;
  block_number: number;
  error_message: string;
  golden_dog_score: number;
  decision_reason: string;
  strategy_used: string;
  current_value: string;
  profit_loss: number;
  user_id: string;
};

export type AITradeStats = {
  count: number;
  winRate: number;
  avgPL: number;
  totalPL: number;
  byStrategy: Array<{
    strategy: string;
    count: number;
    winRate: number;
    avgPL: number;
    totalPL: number;
  }>;
  byPeriod: Array<{
    period: string;
    count: number;
    winRate: number;
    avgPL: number;
    totalPL: number;
  }>;
};
