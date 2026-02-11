import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";

export type RiskFactors = {
  honeypotRisk?: "LOW" | "MEDIUM" | "HIGH";
  taxRisk?: "LOW" | "MEDIUM" | "HIGH";
  ownerRisk?: "LOW" | "MEDIUM" | "HIGH";
  concentrationRisk?: "LOW" | "MEDIUM" | "HIGH";
};

export type RuleWeights = {
  baseMultiplier: number;
  goldenDogBias: number;
  highPenalty: number;
  mediumPenalty: number;
};

export type MemoryState = {
  version: number;
  updatedAt: string;
  weights: RuleWeights;
  outcomes: Array<{
    tokenAddress: string;
    outcome: "MOON" | "RUG" | "FLAT";
    maxGain?: number;
    maxLoss?: number;
    isGoldenDog?: boolean;
    riskFactors?: RiskFactors;
    confidenceWeight?: number;
    timestamp: string;
  }>;
  feedbacks?: UserFeedback[];
  userReputations?: UserReputation[];
  rulePerformance?: RulePerformance[];
};

const DEFAULT_WEIGHTS: RuleWeights = {
  baseMultiplier: 1.0,
  goldenDogBias: 12,
  highPenalty: 15,
  mediumPenalty: 6,
};

export type UserFeedback = {
  tokenAddress: string;
  feedbackType: "CONFIRM_GOLDEN" | "DENY_GOLDEN" | "REPORT_RUG";
  userId: string;
  channel: "OPENCLAW_DIALOG" | "TELEGRAM";
  userReputation: number;
  feedbackWeight: number;
  timestamp: string;
};

export type UserReputation = {
  userId: string;
  reputation: number;
  feedbackCount: number;
  lastSeenAt: string;
};

export type RulePerformance = {
  ruleId: string;
  correct: number;
  total: number;
  accuracy: number;
  updatedAt: string;
};

export type PerformanceWindow = {
  window: "7d" | "30d" | "all";
  byRule: RulePerformance[];
};

function resolveMemoryPath() {
  const explicit = process.env.EASYMEME_MEMORY_PATH?.trim();
  if (explicit) {
    return explicit;
  }
  return path.join(os.homedir(), ".easymeme", "memory.json");
}

async function ensureDir(filePath: string) {
  const dir = path.dirname(filePath);
  await fs.mkdir(dir, { recursive: true });
}

export async function loadMemory(): Promise<MemoryState> {
  const filePath = resolveMemoryPath();
  try {
    const raw = await fs.readFile(filePath, "utf-8");
    const parsed = JSON.parse(raw) as MemoryState;
    if (!parsed?.weights) {
      throw new Error("invalid memory");
    }
    return parsed;
  } catch {
    return {
      version: 1,
      updatedAt: new Date().toISOString(),
      weights: { ...DEFAULT_WEIGHTS },
      outcomes: [],
      feedbacks: [],
      userReputations: [],
      rulePerformance: [],
    };
  }
}

export async function saveMemory(state: MemoryState): Promise<void> {
  const filePath = resolveMemoryPath();
  await ensureDir(filePath);
  await fs.writeFile(filePath, JSON.stringify(state, null, 2), "utf-8");
}

export function countRiskLevels(factors?: RiskFactors) {
  const list = factors ? Object.values(factors) : [];
  let high = 0;
  let medium = 0;
  for (const item of list) {
    if (item === "HIGH") {
      high += 1;
    } else if (item === "MEDIUM") {
      medium += 1;
    }
  }
  return { high, medium };
}

export function estimateScore(
  weights: RuleWeights,
  input: {
    riskScore: number;
    isGoldenDog: boolean;
    riskFactors?: RiskFactors;
  }
) {
  const { high, medium } = countRiskLevels(input.riskFactors);
  const bias = input.isGoldenDog ? weights.goldenDogBias : -Math.round(weights.goldenDogBias / 2);
  const raw =
    input.riskScore * weights.baseMultiplier +
    bias -
    weights.highPenalty * high -
    weights.mediumPenalty * medium;
  const clamped = Math.max(0, Math.min(100, Math.round(raw)));
  return clamped;
}

export function updateWeights(
  weights: RuleWeights,
  input: {
    outcome: "MOON" | "RUG" | "FLAT";
    isGoldenDog?: boolean;
  }
): RuleWeights {
  const next = { ...weights };
  const gd = input.isGoldenDog ?? false;
  if (input.outcome === "MOON" && gd) {
    next.goldenDogBias = Math.min(25, next.goldenDogBias + 1);
    next.highPenalty = Math.max(8, next.highPenalty - 0.5);
    next.mediumPenalty = Math.max(4, next.mediumPenalty - 0.25);
  } else if (input.outcome === "RUG" && gd) {
    next.goldenDogBias = Math.max(4, next.goldenDogBias - 2);
    next.highPenalty = Math.min(25, next.highPenalty + 1);
    next.mediumPenalty = Math.min(12, next.mediumPenalty + 0.5);
  } else if (input.outcome === "FLAT" && gd) {
    next.goldenDogBias = Math.max(6, next.goldenDogBias - 0.5);
  }
  return next;
}

export function applyFeedback(
  weights: RuleWeights,
  feedback: {
    feedbackType: "CONFIRM_GOLDEN" | "DENY_GOLDEN" | "REPORT_RUG";
    weight: number;
  }
): RuleWeights {
  const next = { ...weights };
  const w = Math.max(0, Math.min(1, feedback.weight));
  switch (feedback.feedbackType) {
    case "CONFIRM_GOLDEN":
      next.goldenDogBias = Math.min(25, next.goldenDogBias + 1.2 * w);
      break;
    case "DENY_GOLDEN":
      next.goldenDogBias = Math.max(4, next.goldenDogBias - 1.5 * w);
      next.mediumPenalty = Math.min(12, next.mediumPenalty + 0.4 * w);
      break;
    case "REPORT_RUG":
      next.goldenDogBias = Math.max(2, next.goldenDogBias - 2.0 * w);
      next.highPenalty = Math.min(25, next.highPenalty + 0.8 * w);
      next.mediumPenalty = Math.min(12, next.mediumPenalty + 0.5 * w);
      break;
  }
  return next;
}

export function decayFeedbackWeight(
  baseWeight: number,
  userFeedbackCount: number,
  userReputation: number,
): number {
  const clampedBase = Math.max(0, Math.min(1, baseWeight));
  const count = Math.max(1, userFeedbackCount);
  const countDecay = 1 / (1 + Math.log10(count));
  const repBoost = 0.7 + Math.max(0, Math.min(100, userReputation)) / 100 * 0.3;
  const decayed = clampedBase * countDecay * repBoost;
  return Math.max(0.05, Math.min(1, decayed));
}

export function upsertUserReputation(
  list: UserReputation[] | undefined,
  input: { userId: string; reputation: number }
): UserReputation[] {
  const next = Array.isArray(list) ? [...list] : [];
  const idx = next.findIndex((entry) => entry.userId === input.userId);
  const now = new Date().toISOString();
  if (idx >= 0) {
    next[idx] = {
      ...next[idx],
      reputation: input.reputation,
      feedbackCount: next[idx].feedbackCount + 1,
      lastSeenAt: now
    };
    return next;
  }
  next.push({
    userId: input.userId,
    reputation: input.reputation,
    feedbackCount: 1,
    lastSeenAt: now
  });
  return next;
}

export function updateRulePerformanceOnOutcome(
  list: RulePerformance[] | undefined,
  input: { ruleId: string; outcome: "MOON" | "RUG" | "FLAT"; isGoldenDog?: boolean }
): RulePerformance[] {
  const next = Array.isArray(list) ? [...list] : [];
  if (!input.isGoldenDog && input.isGoldenDog !== false) {
    return next;
  }
  if (input.outcome === "FLAT") {
    return next;
  }
  const correct =
    (input.isGoldenDog && input.outcome === "MOON") ||
    (!input.isGoldenDog && input.outcome === "RUG");
  const now = new Date().toISOString();
  const idx = next.findIndex((entry) => entry.ruleId === input.ruleId);
  if (idx >= 0) {
    const updated = { ...next[idx] };
    updated.total += 1;
    if (correct) {
      updated.correct += 1;
    }
    updated.accuracy = updated.total > 0 ? updated.correct / updated.total : 0;
    updated.updatedAt = now;
    next[idx] = updated;
    return next;
  }
  next.push({
    ruleId: input.ruleId,
    correct: correct ? 1 : 0,
    total: 1,
    accuracy: correct ? 1 : 0,
    updatedAt: now
  });
  return next;
}

function isFactorPredictionCorrect(
  level: "LOW" | "MEDIUM" | "HIGH" | undefined,
  outcome: "MOON" | "RUG" | "FLAT",
): boolean | null {
  if (!level || outcome === "FLAT") {
    return null;
  }
  if (outcome === "RUG") {
    return level === "HIGH" || level === "MEDIUM";
  }
  return level === "LOW";
}

export function updateFactorPerformanceOnOutcome(
  list: RulePerformance[] | undefined,
  input: {
    outcome: "MOON" | "RUG" | "FLAT";
    riskFactors?: RiskFactors;
    confidenceWeight?: number;
  },
): RulePerformance[] {
  const next = Array.isArray(list) ? [...list] : [];
  if (input.outcome === "FLAT" || !input.riskFactors) {
    return next;
  }
  const now = new Date().toISOString();
  const weight = Math.max(0.1, Math.min(1, input.confidenceWeight ?? 1));
  const entries: Array<[string, "LOW" | "MEDIUM" | "HIGH" | undefined]> = [
    ["factor_honeypot", input.riskFactors.honeypotRisk],
    ["factor_tax", input.riskFactors.taxRisk],
    ["factor_owner", input.riskFactors.ownerRisk],
    ["factor_concentration", input.riskFactors.concentrationRisk],
  ];
  for (const [ruleId, level] of entries) {
    const correct = isFactorPredictionCorrect(level, input.outcome);
    if (correct === null) {
      continue;
    }
    const idx = next.findIndex((entry) => entry.ruleId === ruleId);
    if (idx >= 0) {
      const updated = { ...next[idx] };
      updated.total += weight;
      if (correct) {
        updated.correct += weight;
      }
      updated.accuracy = updated.total > 0 ? updated.correct / updated.total : 0;
      updated.updatedAt = now;
      next[idx] = updated;
      continue;
    }
    next.push({
      ruleId,
      correct: correct ? weight : 0,
      total: weight,
      accuracy: correct ? 1 : 0,
      updatedAt: now
    });
  }
  return next;
}

export function buildPerformanceWindows(state: MemoryState): PerformanceWindow[] {
  const outcomes = Array.isArray(state.outcomes) ? state.outcomes : [];
  const now = Date.now();
  const build = (days: number | null, window: "7d" | "30d" | "all"): PerformanceWindow => {
    const from = days === null ? 0 : now - days * 24 * 60 * 60 * 1000;
    let rulePerf: RulePerformance[] = [];
    for (const out of outcomes) {
      const ts = Date.parse(out.timestamp || "");
      if (Number.isNaN(ts) || ts < from) {
        continue;
      }
      rulePerf = updateRulePerformanceOnOutcome(rulePerf, {
        ruleId: "golden_dog_decision",
        outcome: out.outcome,
        isGoldenDog: out.isGoldenDog
      });
      rulePerf = updateFactorPerformanceOnOutcome(rulePerf, {
        outcome: out.outcome,
        riskFactors: out.riskFactors,
        confidenceWeight: out.confidenceWeight
      });
    }
    return { window, byRule: rulePerf };
  };
  return [
    build(7, "7d"),
    build(30, "30d"),
    build(null, "all"),
  ];
}
