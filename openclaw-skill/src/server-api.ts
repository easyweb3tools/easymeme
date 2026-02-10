import type { AIPosition, AITradePayload, PendingToken, TokenRiskAnalysis } from "./types.js";
import crypto from "node:crypto";

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

function signRequest(method: string, path: string, body: string): Record<string, string> {
  const secret = process.env.EASYMEME_API_HMAC_SECRET?.trim();
  if (!secret) {
    return {};
  }
  const timestamp = Math.floor(Date.now() / 1000).toString();
  const nonce = crypto.randomUUID();
  const payload = `${method}\n${path}\n${timestamp}\n${nonce}\n${body}`;
  const signature = crypto
    .createHmac("sha256", secret)
    .update(payload)
    .digest("hex");
  return {
    "X-Timestamp": timestamp,
    "X-Nonce": nonce,
    "X-Signature": signature
  };
}

async function requestJson(
  path: string,
  init?: RequestInit,
  overrideUrl?: string,
): Promise<unknown> {
  const base = resolveServerUrl(overrideUrl);
  const url = `${base}${path}`;
  const apiKey = process.env.EASYMEME_API_KEY?.trim();
  const userId = process.env.EASYMEME_USER_ID?.trim();
  const method = (init?.method || "GET").toUpperCase();
  const body = typeof init?.body === "string" ? init.body : "";
  const signatureHeaders = signRequest(method, path, body);
  const res = await fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(apiKey ? { "X-API-Key": apiKey } : {}),
      ...(userId ? { "X-User-Id": userId } : {}),
      ...signatureHeaders,
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
    goplus:
      record.goplus && typeof record.goplus === "object"
        ? (record.goplus as PendingToken["goplus"])
        : undefined,
    dexscreener:
      record.dexscreener && typeof record.dexscreener === "object"
        ? (record.dexscreener as PendingToken["dexscreener"])
        : undefined,
    marketAlerts: Array.isArray(record.marketAlerts)
      ? (record.marketAlerts as PendingToken["marketAlerts"])
      : undefined,
    socialSignals:
      record.socialSignals && typeof record.socialSignals === "object"
        ? (record.socialSignals as PendingToken["socialSignals"])
        : undefined,
    smartMoneySignals:
      record.smartMoneySignals && typeof record.smartMoneySignals === "object"
        ? (record.smartMoneySignals as PendingToken["smartMoneySignals"])
        : undefined,
    contractCode: typeof record.contractCode === "string" ? record.contractCode : undefined,
    holderDistribution:
      record.holderDistribution && typeof record.holderDistribution === "object"
        ? (record.holderDistribution as PendingToken["holderDistribution"])
      : undefined,
    creatorHistory:
      record.creatorHistory && typeof record.creatorHistory === "object"
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

export async function createWallet(userId: string, overrideUrl?: string): Promise<unknown> {
  return requestJson(
    `/api/wallet/create`,
    {
      method: "POST",
      body: JSON.stringify({ userId })
    },
    overrideUrl,
  );
}

export async function getWalletBalance(userId: string, overrideUrl?: string): Promise<unknown> {
  return requestJson(
    `/api/wallet/balance?userId=${encodeURIComponent(userId)}`,
    undefined,
    overrideUrl,
  );
}

export async function upsertWalletConfig(
  userId: string,
  config: Record<string, unknown>,
  overrideUrl?: string,
): Promise<unknown> {
  return requestJson(
    `/api/wallet/config`,
    {
      method: "POST",
      body: JSON.stringify({ userId, config })
    },
    overrideUrl,
  );
}

export async function executeTrade(
  payload: AITradePayload & { userId: string },
  overrideUrl?: string,
): Promise<unknown> {
  const normalizedAmountIn = (() => {
    if (payload.amountIn === undefined || payload.amountIn === null) {
      return undefined;
    }
    if (typeof payload.amountIn === "string") {
      const trimmed = payload.amountIn.trim();
      if (trimmed === "") {
        return undefined;
      }
      const upper = trimmed.toUpperCase();
      if (upper === "ALL" || upper === "100%") {
        return "ALL";
      }
      const ratioMatch = trimmed.match(/^(\d+(?:\.\d+)?)\s*%$/);
      if (ratioMatch) {
        const ratio = Number(ratioMatch[1]) / 100;
        if (Number.isFinite(ratio) && ratio > 0 && ratio <= 1) {
          return ratio.toString();
        }
      }
      return trimmed;
    }
    if (typeof payload.amountIn === "number") {
      return payload.amountIn.toString();
    }
    return undefined;
  })();
  return requestJson(
    `/api/wallet/execute-trade`,
    {
      method: "POST",
      body: JSON.stringify({
        userId: payload.userId,
        tokenAddress: payload.tokenAddress,
        tokenSymbol: payload.tokenSymbol,
        type: payload.type,
        amountIn: normalizedAmountIn,
        amountOut: payload.amountOut,
        goldenDogScore: payload.goldenDogScore,
        decisionReason: payload.decisionReason,
        strategyUsed: payload.strategyUsed,
        profitLoss: payload.profitLoss,
        force: (payload as any).force,
      })
    },
    overrideUrl,
  );
}

export async function getPositions(userId: string, overrideUrl?: string): Promise<AIPosition[]> {
  const payload = await requestJson(
    `/api/ai-positions?userId=${encodeURIComponent(userId)}`,
    undefined,
    overrideUrl
  );
  const list = Array.isArray(payload)
    ? payload
    : payload && typeof payload === "object" && Array.isArray((payload as any).data)
      ? (payload as any).data
      : [];
  return list as AIPosition[];
}
