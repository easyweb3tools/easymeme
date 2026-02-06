import { GoldenDogRecord, RiskPattern, TokenRecord } from "../types";

export interface MemoryStore {
  get: (key: string) => Promise<unknown> | unknown;
  set: (key: string, value: unknown) => Promise<void> | void;
}

export interface ToolContext {
  memory?: MemoryStore;
}

export async function loadTokenHistory(ctx?: ToolContext): Promise<Record<string, TokenRecord>> {
  if (!ctx?.memory) {
    return {};
  }
  const data = await ctx.memory.get("analyzedTokens");
  if (!data || typeof data !== "object") {
    return {};
  }
  return data as Record<string, TokenRecord>;
}

export async function saveTokenHistory(ctx: ToolContext | undefined, history: Record<string, TokenRecord>): Promise<void> {
  if (!ctx?.memory) {
    return;
  }
  await ctx.memory.set("analyzedTokens", history);
}

export async function loadRiskPatterns(ctx?: ToolContext): Promise<Record<string, RiskPattern>> {
  if (!ctx?.memory) {
    return {};
  }
  const data = await ctx.memory.get("riskPatterns");
  if (!data || typeof data !== "object") {
    return {};
  }
  return data as Record<string, RiskPattern>;
}

export async function saveRiskPatterns(ctx: ToolContext | undefined, patterns: Record<string, RiskPattern>): Promise<void> {
  if (!ctx?.memory) {
    return;
  }
  await ctx.memory.set("riskPatterns", patterns);
}

export async function loadGoldenDogHistory(ctx?: ToolContext): Promise<Record<string, GoldenDogRecord>> {
  if (!ctx?.memory) {
    return {};
  }
  const data = await ctx.memory.get("goldenDogHistory");
  if (!data || typeof data !== "object") {
    return {};
  }
  return data as Record<string, GoldenDogRecord>;
}

export async function saveGoldenDogHistory(
  ctx: ToolContext | undefined,
  history: Record<string, GoldenDogRecord>
): Promise<void> {
  if (!ctx?.memory) {
    return;
  }
  await ctx.memory.set("goldenDogHistory", history);
}
