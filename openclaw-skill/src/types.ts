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
  contractCode?: string;
  holderDistribution?: HolderInfo[];
  creatorHistory?: CreatorTx[];
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
  riskFactors: RiskFactors;
  reasoning: string;
  recommendation: string;
};
