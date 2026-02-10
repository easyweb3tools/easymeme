import { Type } from "@sinclair/typebox";
import { jsonResult, readNumberParam, readStringParam } from "openclaw/plugin-sdk";
import {
  createWallet,
  executeTrade,
  fetchPendingTokens,
  getPositions,
  getWalletBalance,
  submitAnalysis,
  upsertWalletConfig
} from "./server-api.js";
import { notifyGoldenDogFound, notifySellTrade } from "./notify.js";
import {
  applyFeedback,
  estimateScore,
  loadMemory,
  saveMemory,
  updateWeights,
  updateRulePerformanceOnOutcome,
  upsertUserReputation
} from "./memory.js";

type AnyAgentTool = any;
const HolderInfoSchema = Type.Object(
  {
    address: Type.Optional(Type.String()),
    balance: Type.Optional(Type.Number()),
    percentage: Type.Optional(Type.Number()),
    isContract: Type.Optional(Type.Boolean())
  },
  { additionalProperties: true }
);

const CreatorTxSchema = Type.Object(
  {
    hash: Type.Optional(Type.String()),
    to: Type.Optional(Type.String()),
    from: Type.Optional(Type.String()),
    value: Type.Optional(Type.String()),
    timestamp: Type.Optional(Type.String())
  },
  { additionalProperties: true }
);

const TokenSchema = Type.Object({
  address: Type.String(),
  name: Type.Optional(Type.String()),
  symbol: Type.Optional(Type.String()),
  liquidity: Type.Optional(Type.Number()),
  creatorAddress: Type.Optional(Type.String()),
  createdAt: Type.Optional(Type.String()),
  pairAddress: Type.Optional(Type.String()),
  contractCode: Type.Optional(Type.String()),
  holderDistribution: Type.Optional(Type.Array(HolderInfoSchema)),
  creatorHistory: Type.Optional(Type.Array(CreatorTxSchema))
});

const RiskFactorsSchema = Type.Object({
  honeypotRisk: Type.Union([Type.Literal("LOW"), Type.Literal("MEDIUM"), Type.Literal("HIGH")]),
  taxRisk: Type.Union([Type.Literal("LOW"), Type.Literal("MEDIUM"), Type.Literal("HIGH")]),
  ownerRisk: Type.Union([Type.Literal("LOW"), Type.Literal("MEDIUM"), Type.Literal("HIGH")]),
  concentrationRisk: Type.Union([
    Type.Literal("LOW"),
    Type.Literal("MEDIUM"),
    Type.Literal("HIGH")
  ])
});

const AnalysisSchema = Type.Object({
  riskScore: Type.Number(),
  riskLevel: Type.Union([
    Type.Literal("SAFE"),
    Type.Literal("WARNING"),
    Type.Literal("DANGER")
  ]),
  isGoldenDog: Type.Boolean(),
  goldenDogScore: Type.Optional(Type.Number()),
  riskFactors: RiskFactorsSchema,
  reasoning: Type.String(),
  recommendation: Type.String()
});

const FetchPendingTokensSchema = Type.Object({
  limit: Type.Optional(Type.Number())
});

const AnalyzeTokenRiskSchema = Type.Object({
  token: TokenSchema,
  analysis: AnalysisSchema
});

const SubmitAnalysisSchema = Type.Object({
  tokenAddress: Type.String(),
  analysis: AnalysisSchema
});

const EstimateGoldenDogScoreSchema = Type.Object({
  riskScore: Type.Number(),
  isGoldenDog: Type.Boolean(),
  riskFactors: Type.Optional(RiskFactorsSchema)
});

const RecordOutcomeSchema = Type.Object({
  tokenAddress: Type.String(),
  outcome: Type.Union([Type.Literal("MOON"), Type.Literal("RUG"), Type.Literal("FLAT")]),
  maxGain: Type.Optional(Type.Number()),
  maxLoss: Type.Optional(Type.Number()),
  lessonsLearned: Type.Optional(Type.String()),
  analysis: Type.Optional(
    Type.Object({
      isGoldenDog: Type.Optional(Type.Boolean())
    })
  )
});

const ExecuteTradeSchema = Type.Object({
  tokenAddress: Type.Optional(Type.String()),
  tokenSymbol: Type.Optional(Type.String()),
  type: Type.Union([Type.Literal("BUY"), Type.Literal("SELL")]),
  amountIn: Type.Optional(
    Type.Union([Type.String(), Type.Number()])
  ),
  amountOut: Type.Optional(Type.String()),
  goldenDogScore: Type.Optional(Type.Number()),
  decisionReason: Type.Optional(Type.String()),
  strategyUsed: Type.Optional(Type.String()),
  userId: Type.Optional(Type.String()),
  profitLoss: Type.Optional(Type.Number()),
  force: Type.Optional(Type.Boolean())
});

const WalletInfoSchema = Type.Object({
  userId: Type.Optional(Type.String())
});

const PositionsSchema = Type.Object({
  userId: Type.Optional(Type.String()),
  format: Type.Optional(Type.Union([Type.Literal("summary"), Type.Literal("detailed")]))
});

const WalletConfigSchema = Type.Object({
  userId: Type.Optional(Type.String()),
  config: Type.Optional(Type.Object({}, { additionalProperties: true }))
});

const RecordUserFeedbackSchema = Type.Object({
  tokenAddress: Type.String(),
  feedbackType: Type.Union([
    Type.Literal("CONFIRM_GOLDEN"),
    Type.Literal("DENY_GOLDEN"),
    Type.Literal("REPORT_RUG")
  ]),
  userId: Type.String(),
  channel: Type.Union([Type.Literal("OPENCLAW_DIALOG"), Type.Literal("TELEGRAM")]),
  userReputation: Type.Optional(Type.Number())
});

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}

function validateAnalysis(value: unknown): asserts value is Record<string, unknown> {
  if (!isRecord(value)) {
    throw new Error("analysis must be an object");
  }
  const required = [
    "riskScore",
    "riskLevel",
    "isGoldenDog",
    "riskFactors",
    "reasoning",
    "recommendation"
  ];
  for (const key of required) {
    if (!(key in value)) {
      throw new Error(`analysis.${key} is required`);
    }
  }
  if (!isRecord(value.riskFactors)) {
    throw new Error("analysis.riskFactors is required");
  }
}

async function ensureGoldenDogScore(analysis: Record<string, unknown>) {
  if (typeof analysis.goldenDogScore === "number") {
    return;
  }
  const memory = await loadMemory();
  const riskScore =
    typeof analysis.riskScore === "number" ? analysis.riskScore : 0;
  const isGoldenDog = Boolean(analysis.isGoldenDog);
  const riskFactors = analysis.riskFactors as any;
  analysis.goldenDogScore = estimateScore(memory.weights, {
    riskScore,
    isGoldenDog,
    riskFactors
  });
}

export function createFetchPendingTokensTool(options?: { serverUrl?: string }): AnyAgentTool {
  return {
    label: "Fetch Pending Tokens",
    name: "fetchPendingTokens",
    description: "Fetch tokens pending analysis from the EasyMeme server API.",
    parameters: FetchPendingTokensSchema,
    execute: async (_toolCallId: string, params: Record<string, unknown>) => {
      const limit = readNumberParam(params, "limit") ?? 10;
      const tokens = await fetchPendingTokens(
        Math.max(1, Math.trunc(limit)),
        options?.serverUrl,
      );
      return jsonResult({ tokens, count: tokens.length });
    }
  };
}

export function createAnalyzeTokenRiskTool(): AnyAgentTool {
  return {
    label: "Analyze Token Risk",
    name: "analyzeTokenRisk",
    description:
      "Record an AI-generated risk analysis for a token. Use the model to produce the analysis, then call this tool with the structured result.",
    parameters: AnalyzeTokenRiskSchema,
    execute: async (_toolCallId: string, params: Record<string, unknown>) => {
      const token = params.token as unknown;
      const analysis = params.analysis as unknown;
      validateAnalysis(analysis);
      await ensureGoldenDogScore(analysis as Record<string, unknown>);
      return jsonResult({ ok: true, token, analysis });
    }
  };
}

export function createSubmitAnalysisTool(options?: { serverUrl?: string }): AnyAgentTool {
  return {
    label: "Submit Analysis",
    name: "submitAnalysis",
    description: "Submit a completed token analysis back to the EasyMeme server.",
    parameters: SubmitAnalysisSchema,
    execute: async (_toolCallId: string, params: Record<string, unknown>) => {
      const tokenAddress = readStringParam(params, "tokenAddress", { required: true });
      const analysis = params.analysis as unknown;
      validateAnalysis(analysis);
      await ensureGoldenDogScore(analysis as Record<string, unknown>);
      const analysisRecord = analysis as Record<string, unknown>;
      const result = await submitAnalysis(tokenAddress, analysis as any, options?.serverUrl);

      if (analysisRecord.isGoldenDog === true) {
        await notifyGoldenDogFound({
          tokenAddress,
          tokenSymbol: readStringParam(params, "tokenSymbol") || undefined,
          goldenDogScore:
            typeof analysisRecord.goldenDogScore === "number"
              ? analysisRecord.goldenDogScore
              : undefined,
          riskScore:
            typeof analysisRecord.riskScore === "number"
              ? analysisRecord.riskScore
              : undefined,
          decisionReason:
            typeof analysisRecord.decisionReason === "string"
              ? analysisRecord.decisionReason
              : undefined,
        });
      }

      return jsonResult({ ok: true, result });
    }
  };
}

export function createEstimateGoldenDogScoreTool(): AnyAgentTool {
  return {
    label: "Estimate Golden Dog Score",
    name: "estimateGoldenDogScore",
    description:
      "Estimate golden dog score using locally learned weights stored in OpenClaw memory.",
    parameters: EstimateGoldenDogScoreSchema,
    execute: async (_toolCallId: string, params: Record<string, unknown>) => {
      const memory = await loadMemory();
      const riskScore = readNumberParam(params, "riskScore");
      if (typeof riskScore !== "number") {
        throw new Error("riskScore is required");
      }
      const isGoldenDog = Boolean(params.isGoldenDog);
      const riskFactors = params.riskFactors as any;
      const score = estimateScore(memory.weights, { riskScore, isGoldenDog, riskFactors });
      return jsonResult({ score, weights: memory.weights });
    }
  };
}

export function createRecordOutcomeTool(): AnyAgentTool {
  return {
    label: "Record Outcome",
    name: "recordOutcome",
    description:
      "Record a trade outcome and update learned weights stored in OpenClaw memory.",
    parameters: RecordOutcomeSchema,
    execute: async (_toolCallId: string, params: Record<string, unknown>) => {
      const tokenAddress = readStringParam(params, "tokenAddress", { required: true });
      const outcome = readStringParam(params, "outcome", { required: true }) as
        | "MOON"
        | "RUG"
        | "FLAT";
      const maxGain = readNumberParam(params, "maxGain");
      const maxLoss = readNumberParam(params, "maxLoss");
      const analysis = params.analysis as { isGoldenDog?: boolean } | undefined;

      const memory = await loadMemory();
      memory.outcomes.push({
        tokenAddress,
        outcome,
        maxGain: typeof maxGain === "number" ? maxGain : undefined,
        maxLoss: typeof maxLoss === "number" ? maxLoss : undefined,
        timestamp: new Date().toISOString()
      });
      memory.weights = updateWeights(memory.weights, {
        outcome,
        isGoldenDog: analysis?.isGoldenDog
      });
      memory.rulePerformance = updateRulePerformanceOnOutcome(memory.rulePerformance, {
        ruleId: "golden_dog_decision",
        outcome,
        isGoldenDog: analysis?.isGoldenDog
      });
      memory.updatedAt = new Date().toISOString();
      await saveMemory(memory);

      return jsonResult({
        ok: true,
        weights: memory.weights,
        rulePerformance: memory.rulePerformance
      });
    }
  };
}

export function createExecuteTradeTool(options?: { serverUrl?: string; userId?: string }): AnyAgentTool {
  return {
    label: "Execute Trade",
    name: "executeTrade",
    description:
      "Execute an auto trade using managed wallet and record it as an AI trade.",
    parameters: ExecuteTradeSchema,
    execute: async (_toolCallId: string, params: Record<string, unknown>) => {
      const tokenAddress = readStringParam(params, "tokenAddress");
      const tokenSymbol = readStringParam(params, "tokenSymbol");
      const type = readStringParam(params, "type", { required: true }) as "BUY" | "SELL";
      const amountInRaw = params.amountIn as unknown;
      let amountIn: string | undefined;
      if (typeof amountInRaw === "number") {
        amountIn = amountInRaw.toString();
      } else if (typeof amountInRaw === "string") {
        amountIn = amountInRaw.trim();
      }
      const amountOut = readStringParam(params, "amountOut");
      const decisionReason = readStringParam(params, "decisionReason");
      const strategyUsed = readStringParam(params, "strategyUsed");
      const goldenDogScore = readNumberParam(params, "goldenDogScore");
      const profitLoss = readNumberParam(params, "profitLoss");
      const force = Boolean(params.force);
      const userId =
        readStringParam(params, "userId") ||
        options?.userId ||
        process.env.EASYMEME_USER_ID ||
        "default";

      try {
        await getWalletBalance(userId, options?.serverUrl);
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        if (message.includes("404")) {
          await createWallet(userId, options?.serverUrl);
        } else {
          throw err;
        }
      }

      let resolvedTokenAddress = tokenAddress;
      if (!resolvedTokenAddress && tokenSymbol) {
        const positions = await getPositions(userId, options?.serverUrl);
        const match = positions.find(
          (pos) =>
            (pos.token_symbol || "").toLowerCase() === tokenSymbol.toLowerCase()
        );
        if (match?.token_address) {
          resolvedTokenAddress = match.token_address;
        } else {
          throw new Error(`tokenSymbol not found in positions: ${tokenSymbol}`);
        }
      }
      if (!resolvedTokenAddress) {
        throw new Error("tokenAddress is required");
      }

      const result = await executeTrade(
        {
          userId,
          tokenAddress: resolvedTokenAddress,
          tokenSymbol: tokenSymbol || undefined,
          type,
          amountIn: amountIn || undefined,
          amountOut: amountOut || undefined,
          goldenDogScore: typeof goldenDogScore === "number" ? goldenDogScore : undefined,
          decisionReason: decisionReason || undefined,
          strategyUsed: strategyUsed || undefined,
          profitLoss: typeof profitLoss === "number" ? profitLoss : undefined,
          force: force ? true : undefined,
        },
        options?.serverUrl,
      );

      if (type === "SELL") {
        await notifySellTrade(
          {
            tokenAddress: resolvedTokenAddress,
            tokenSymbol: tokenSymbol || undefined,
            amountIn: amountIn || undefined
          },
          result
        );
      }

      return jsonResult({ ok: true, result });
    }
  };
}

export function createUpsertWalletConfigTool(options?: { serverUrl?: string; userId?: string }): AnyAgentTool {
  return {
    label: "Upsert Wallet Config",
    name: "upsertWalletConfig",
    description: "Update managed wallet auto-trade config.",
    parameters: WalletConfigSchema,
    execute: async (_toolCallId: string, params: Record<string, unknown>) => {
      const userId =
        readStringParam(params, "userId") ||
        options?.userId ||
        process.env.EASYMEME_USER_ID ||
        "default";
      const config =
        (params.config as Record<string, unknown> | undefined) ?? {};
      const result = await upsertWalletConfig(userId, config, options?.serverUrl);
      return jsonResult({ ok: true, result });
    }
  };
}

export function createGetWalletInfoTool(options?: { serverUrl?: string; userId?: string }): AnyAgentTool {
  return {
    label: "Get Wallet Info",
    name: "getWalletInfo",
    description: "Fetch managed wallet address and balance for deposits.",
    parameters: WalletInfoSchema,
    execute: async (_toolCallId: string, params: Record<string, unknown>) => {
      const userId =
        readStringParam(params, "userId") ||
        options?.userId ||
        process.env.EASYMEME_USER_ID ||
        "default";
      const result = await getWalletBalance(userId, options?.serverUrl);
      return jsonResult({ ok: true, result });
    }
  };
}

export function createGetPositionsTool(options?: { serverUrl?: string; userId?: string }): AnyAgentTool {
  return {
    label: "Get AI Positions",
    name: "getPositions",
    description: "Fetch current AI positions for a user to decide sell amounts.",
    parameters: PositionsSchema,
    execute: async (_toolCallId: string, params: Record<string, unknown>) => {
      const userId =
        readStringParam(params, "userId") ||
        options?.userId ||
        process.env.EASYMEME_USER_ID ||
        "default";
      const format = readStringParam(params, "format") || "summary";
      const positions = await getPositions(userId, options?.serverUrl);
      if (format === "detailed") {
        return jsonResult({ ok: true, positions, count: positions.length });
      }
      const summary = positions.map((pos) => {
        const qty = pos.quantity ?? "0";
        const cost = pos.cost_bnb ?? "0";
        const symbol = pos.token_symbol || "UNKNOWN";
        const updated = pos.updated_at || "unknown";
        return `${symbol} | ${pos.token_address} | qty=${qty} | cost=${cost} | updated=${updated}`;
      });
      return jsonResult({ ok: true, positions: summary, count: summary.length });
    }
  };
}

export function createRecordUserFeedbackTool(): AnyAgentTool {
  return {
    label: "Record User Feedback",
    name: "recordUserFeedback",
    description:
      "Record user feedback from OpenClaw Dialog or Telegram and update local memory weights.",
    parameters: RecordUserFeedbackSchema,
    execute: async (_toolCallId: string, params: Record<string, unknown>) => {
      const tokenAddress = readStringParam(params, "tokenAddress", { required: true });
      const feedbackType = readStringParam(params, "feedbackType", { required: true }) as
        | "CONFIRM_GOLDEN"
        | "DENY_GOLDEN"
        | "REPORT_RUG";
      const userId = readStringParam(params, "userId", { required: true });
      const channel = readStringParam(params, "channel", { required: true }) as
        | "OPENCLAW_DIALOG"
        | "TELEGRAM";
      const rep = readNumberParam(params, "userReputation");
      const reputation = typeof rep === "number" ? Math.max(0, Math.min(100, rep)) : 30;
      const weight = reputation / 100;

      const memory = await loadMemory();
      const feedback = {
        tokenAddress,
        feedbackType,
        userId,
        channel,
        userReputation: reputation,
        feedbackWeight: weight,
        timestamp: new Date().toISOString()
      };
      memory.feedbacks = Array.isArray(memory.feedbacks) ? memory.feedbacks : [];
      memory.feedbacks.push(feedback);
      memory.userReputations = upsertUserReputation(memory.userReputations, {
        userId,
        reputation
      });
      memory.weights = applyFeedback(memory.weights, {
        feedbackType,
        weight
      });
      memory.updatedAt = new Date().toISOString();
      await saveMemory(memory);

      return jsonResult({ ok: true, weights: memory.weights, feedback });
    }
  };
}
