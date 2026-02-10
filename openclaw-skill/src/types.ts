export type HolderInfo = {
  address?: string;
  balance?: number;
  percentage?: number;
  isContract?: boolean;
  [key: string]: unknown;
};

export type CreatorTx = {
  hash?: string;
  to?: string;
  from?: string;
  value?: string;
  timestamp?: string;
  [key: string]: unknown;
};

export type PendingToken = {
  address: string;
  name?: string;
  symbol?: string;
  liquidity?: number;
  creatorAddress?: string;
  createdAt?: string;
  pairAddress?: string;
  goplus?: Record<string, unknown>;
  dexscreener?: Record<string, unknown>;
  marketAlerts?: Record<string, unknown>[];
  socialSignals?: Record<string, unknown>;
  smartMoneySignals?: Record<string, unknown>;
  contractCode?: string;
  holderDistribution?: Record<string, unknown> | HolderInfo[];
  creatorHistory?: Record<string, unknown> | CreatorTx[];
  [key: string]: unknown;
};

export type RiskFactors = {
  honeypotRisk: "LOW" | "MEDIUM" | "HIGH";
  taxRisk: "LOW" | "MEDIUM" | "HIGH";
  ownerRisk: "LOW" | "MEDIUM" | "HIGH";
  concentrationRisk: "LOW" | "MEDIUM" | "HIGH";
};

export type TokenRiskAnalysis = {
  riskScore: number;
  riskLevel: "SAFE" | "WARNING" | "DANGER";
  isGoldenDog: boolean;
  goldenDogScore?: number;
  riskFactors: RiskFactors;
  reasoning: string;
  recommendation: string;
};

export type AITradePayload = {
  tokenAddress: string;
  tokenSymbol?: string;
  type: "BUY" | "SELL";
  amountIn?: string | number;
  amountOut?: string;
  txHash?: string;
  goldenDogScore?: number;
  decisionReason?: string;
  strategyUsed?: string;
  currentValue?: string;
  profitLoss?: number;
  force?: boolean;
};

export type AIPosition = {
  user_id: string;
  token_address: string;
  token_symbol?: string;
  quantity?: string;
  cost_bnb?: string;
  updated_at?: string;
};
