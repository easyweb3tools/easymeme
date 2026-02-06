export interface FetchPendingTokensInput {
  limit?: number;
}

export interface HolderInfo {
  address: string;
  percent: number;
}

export interface CreatorTx {
  hash: string;
  action: string;
  timestamp: string;
}

export interface PendingToken {
  address: string;
  name: string;
  symbol: string;
  liquidity: number;
  creatorAddress: string;
  createdAt?: string;
  pairAddress?: string;
  contractCode?: string;
  holderDistribution?: HolderInfo[];
  creatorHistory?: CreatorTx[];
}

export type RiskLevel = "SAFE" | "WARNING" | "DANGER";
export type RiskBand = "LOW" | "MEDIUM" | "HIGH";

export interface AnalyzeTokenRiskInput {
  token: PendingToken;
}

export interface AnalyzeTokenRiskOutput {
  riskScore: number;
  riskLevel: RiskLevel;
  isGoldenDog: boolean;
  riskFactors: {
    honeypotRisk: RiskBand;
    taxRisk: RiskBand;
    ownerRisk: RiskBand;
    concentrationRisk: RiskBand;
  };
  reasoning: string;
  recommendation: string;
}

export interface SubmitAnalysisInput {
  tokenAddress: string;
  analysis: AnalyzeTokenRiskOutput;
}

export interface TokenRecord {
  tokenAddress: string;
  firstSeenAt: string;
  lastAnalyzedAt: string;
  analysisHistory: AnalyzeTokenRiskOutput[];
  userInterest: boolean;
}

export interface RiskPattern {
  patternHash: string;
  description: string;
  count: number;
}

export interface GoldenDogRecord {
  tokenAddress: string;
  detectedAt: string;
  reasoning: string;
}
