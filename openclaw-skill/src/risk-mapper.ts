import type { PendingToken, RiskFactors, TokenRiskAnalysis } from "./types.js";

const LEVELS = ["LOW", "MEDIUM", "HIGH"] as const;

type RiskLevel = typeof LEVELS[number];

function toBool(v: unknown): boolean {
  if (typeof v === "boolean") {
    return v;
  }
  if (typeof v === "number") {
    return v !== 0;
  }
  if (typeof v === "string") {
    const normalized = v.trim().toLowerCase();
    return normalized === "1" || normalized === "true" || normalized === "yes";
  }
  return false;
}

function toNumber(v: unknown): number {
  if (typeof v === "number" && Number.isFinite(v)) {
    return v;
  }
  if (typeof v === "string") {
    const parsed = Number(v);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  return 0;
}

function readNested(obj: Record<string, unknown> | undefined, path: string[]): unknown {
  if (!obj) {
    return undefined;
  }
  let current: unknown = obj;
  for (const key of path) {
    if (!current || typeof current !== "object") {
      return undefined;
    }
    current = (current as Record<string, unknown>)[key];
  }
  return current;
}

function normalizeTax(tax: number): number {
  if (tax <= 1) {
    return tax;
  }
  return tax / 100;
}

function classifyTaxRisk(buyTax: number, sellTax: number): RiskLevel {
  const maxTax = Math.max(normalizeTax(buyTax), normalizeTax(sellTax));
  if (maxTax >= 0.15) {
    return "HIGH";
  }
  if (maxTax >= 0.08) {
    return "MEDIUM";
  }
  return "LOW";
}

function classifyOwnerRisk(goplus: Record<string, unknown>): RiskLevel {
  const mintable = toBool(goplus.is_mintable);
  const takeBack = toBool(goplus.can_take_back_ownership);
  const proxy = toBool(goplus.is_proxy);
  if (mintable || takeBack) {
    return "HIGH";
  }
  if (proxy) {
    return "MEDIUM";
  }
  return "LOW";
}

function classifyConcentrationRisk(
  holderDistribution?: Record<string, unknown>,
  dexscreener?: Record<string, unknown>,
): RiskLevel {
  const top10Share = toNumber(holderDistribution?.top10Share);
  if (top10Share >= 0.8) {
    return "HIGH";
  }
  if (top10Share >= 0.6) {
    return "MEDIUM";
  }

  const buysH1 = toNumber(readNested(dexscreener, ["txns", "h1", "buys"]));
  const sellsH1 = toNumber(readNested(dexscreener, ["txns", "h1", "sells"]));
  if (sellsH1 > buysH1 * 2 && sellsH1 >= 20) {
    return "MEDIUM";
  }
  return "LOW";
}

function scoreFromFactors(riskFactors: RiskFactors): number {
  let score = 80;
  for (const key of Object.keys(riskFactors) as Array<keyof RiskFactors>) {
    const level = riskFactors[key];
    if (level === "HIGH") {
      score -= 30;
    } else if (level === "MEDIUM") {
      score -= 12;
    }
  }
  if (score < 0) {
    return 0;
  }
  if (score > 100) {
    return 100;
  }
  return score;
}

function classifyRiskLevel(riskScore: number): "SAFE" | "WARNING" | "DANGER" {
  if (riskScore >= 70) {
    return "SAFE";
  }
  if (riskScore >= 45) {
    return "WARNING";
  }
  return "DANGER";
}

export function buildAnalysisDraft(token: PendingToken): TokenRiskAnalysis {
  const goplus = (token.goplus ?? {}) as Record<string, unknown>;
  const dexscreener = (token.dexscreener ?? {}) as Record<string, unknown>;
  const holderDistribution = token.holderDistribution && typeof token.holderDistribution === "object"
    ? (token.holderDistribution as Record<string, unknown>)
    : undefined;

  const honeypotRisk: RiskLevel = toBool(goplus.is_honeypot) ? "HIGH" : "LOW";
  const taxRisk = classifyTaxRisk(toNumber(goplus.buy_tax), toNumber(goplus.sell_tax));
  const ownerRisk = classifyOwnerRisk(goplus);
  const concentrationRisk = classifyConcentrationRisk(holderDistribution, dexscreener);

  const riskFactors: RiskFactors = {
    honeypotRisk,
    taxRisk,
    ownerRisk,
    concentrationRisk,
  };

  const riskScore = scoreFromFactors(riskFactors);
  const riskLevel = classifyRiskLevel(riskScore);

  const priceH1 = toNumber(readNested(dexscreener, ["priceChange", "h1"]));
  const buysH1 = toNumber(readNested(dexscreener, ["txns", "h1", "buys"]));
  const sellsH1 = toNumber(readNested(dexscreener, ["txns", "h1", "sells"]));
  const liqUsd = toNumber(readNested(dexscreener, ["liquidity", "usd"]));

  const momentumGood = priceH1 > 10 && buysH1 >= sellsH1 && liqUsd >= 5000;
  const isGoldenDog = riskLevel !== "DANGER" && honeypotRisk !== "HIGH" && momentumGood;

  const reasoning = [
    `GoPlus honeypot=${String(goplus.is_honeypot ?? "unknown")}, buyTax=${String(goplus.buy_tax ?? "unknown")}, sellTax=${String(goplus.sell_tax ?? "unknown")}`,
    `DEX h1 priceChange=${priceH1}, txns buys/sells=${buysH1}/${sellsH1}, liquidityUsd=${liqUsd}`,
    holderDistribution ? `Holder top10Share=${toNumber(holderDistribution.top10Share)}` : "Holder distribution unavailable",
  ].join(". ");

  const recommendation = isGoldenDog
    ? "Momentum and risk profile are acceptable for a small, controlled position."
    : "Do not auto-buy yet; wait for stronger momentum or better ownership/risk signals.";

  return {
    riskScore,
    riskLevel,
    isGoldenDog,
    riskFactors,
    reasoning,
    recommendation,
  };
}
