import { requireEnv } from "./env";
import { AnalyzeTokenRiskOutput, PendingToken } from "../types";

export interface TokenApiResponse<T> {
  data: T;
}

function baseUrl(): string {
  return requireEnv("SERVER_API_URL").replace(/\/$/, "");
}

function normalizeToken(raw: any): PendingToken {
  return {
    address: String(raw.address ?? raw.token_address ?? ""),
    name: String(raw.name ?? ""),
    symbol: String(raw.symbol ?? ""),
    liquidity: Number(raw.liquidity ?? raw.initial_liquidity ?? 0),
    creatorAddress: String(raw.creatorAddress ?? raw.creator_address ?? ""),
    createdAt: typeof raw.createdAt === "string" ? raw.createdAt : raw.created_at,
    pairAddress: typeof raw.pairAddress === "string" ? raw.pairAddress : raw.pair_address,
    contractCode: typeof raw.contractCode === "string" ? raw.contractCode : raw.contract_code,
    holderDistribution: Array.isArray(raw.holderDistribution)
      ? raw.holderDistribution
      : raw.holder_distribution,
    creatorHistory: Array.isArray(raw.creatorHistory) ? raw.creatorHistory : raw.creator_history
  };
}

export async function fetchPendingTokens(limit: number): Promise<PendingToken[]> {
  const url = new URL(`${baseUrl()}/api/tokens/pending`);
  url.searchParams.set("limit", String(limit));

  const res = await fetch(url.toString());
  if (!res.ok) {
    throw new Error(`Server pending tokens error: ${res.status}`);
  }
  const payload = (await res.json()) as TokenApiResponse<PendingToken[]>;
  const data = payload.data ?? [];
  return data.map(normalizeToken).filter((token) => token.address);
}

export async function fetchToken(address: string): Promise<PendingToken> {
  const res = await fetch(`${baseUrl()}/api/tokens/${address}`);
  if (!res.ok) {
    throw new Error(`Server token error: ${res.status}`);
  }
  const payload = (await res.json()) as TokenApiResponse<PendingToken>;
  return normalizeToken(payload.data);
}

export async function submitAnalysis(tokenAddress: string, analysis: AnalyzeTokenRiskOutput): Promise<void> {
  const res = await fetch(`${baseUrl()}/api/tokens/${tokenAddress}/analysis`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(analysis)
  });
  if (!res.ok) {
    throw new Error(`Server analysis error: ${res.status}`);
  }
}
