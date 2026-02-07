import type { PendingToken, TokenRiskAnalysis } from "./types.js";

const DEFAULT_SERVER_URL = "http://localhost:8080";

function resolveServerUrl(override?: string): string {
  const explicit = override?.trim();
  if (explicit) {
    return explicit.replace(/\/$/, "");
  }
  const env = process.env.EASYMEME_SERVER_URL?.trim();
  if (env) {
    return env.replace(/\/$/, "");
  }
  return DEFAULT_SERVER_URL;
}

async function requestJson(
  path: string,
  init?: RequestInit,
  overrideUrl?: string,
): Promise<unknown> {
  const base = resolveServerUrl(overrideUrl);
  const url = `${base}${path}`;
  const apiKey = process.env.EASYMEME_API_KEY?.trim();
  const res = await fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(apiKey ? { "X-API-Key": apiKey } : {}),
      ...(init?.headers ?? {})
    }
  });
  const text = await res.text();
  if (!res.ok) {
    const suffix = text ? `: ${text}` : "";
    throw new Error(`EasyMeme API ${res.status} ${res.statusText}${suffix}`);
  }
  if (!text) {
    return null;
  }
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return text;
  }
}

function normalizeToken(raw: unknown): PendingToken | null {
  if (!raw || typeof raw !== "object") {
    return null;
  }
  const record = raw as Record<string, unknown>;
  const address = typeof record.address === "string" ? record.address : "";
  if (!address) {
    return null;
  }
  return {
    address,
    name: typeof record.name === "string" ? record.name : undefined,
    symbol: typeof record.symbol === "string" ? record.symbol : undefined,
    liquidity: typeof record.liquidity === "number" ? record.liquidity : undefined,
    creatorAddress:
      typeof record.creatorAddress === "string" ? record.creatorAddress : undefined,
    createdAt: typeof record.createdAt === "string" ? record.createdAt : undefined,
    pairAddress: typeof record.pairAddress === "string" ? record.pairAddress : undefined,
    contractCode: typeof record.contractCode === "string" ? record.contractCode : undefined,
    holderDistribution: Array.isArray(record.holderDistribution)
      ? (record.holderDistribution as PendingToken["holderDistribution"])
      : undefined,
    creatorHistory: Array.isArray(record.creatorHistory)
      ? (record.creatorHistory as PendingToken["creatorHistory"])
      : undefined
  };
}

function normalizeTokenList(payload: unknown): PendingToken[] {
  const list = Array.isArray(payload)
    ? payload
    : payload && typeof payload === "object" && Array.isArray((payload as any).data)
      ? (payload as any).data
      : [];
  const tokens: PendingToken[] = [];
  for (const item of list) {
    const token = normalizeToken(item);
    if (token) {
      tokens.push(token);
    }
  }
  return tokens;
}

export async function fetchPendingTokens(
  limit = 10,
  overrideUrl?: string,
): Promise<PendingToken[]> {
  const payload = await requestJson(
    `/api/tokens/pending?limit=${encodeURIComponent(limit)}`,
    undefined,
    overrideUrl,
  );
  return normalizeTokenList(payload);
}

export async function submitAnalysis(
  tokenAddress: string,
  analysis: TokenRiskAnalysis,
  overrideUrl?: string,
): Promise<unknown> {
  const body = JSON.stringify(analysis);
  return requestJson(
    `/api/tokens/${encodeURIComponent(tokenAddress)}/analysis`,
    {
      method: "POST",
      body,
    },
    overrideUrl,
  );
}
