import { Type } from "@sinclair/typebox";
import { jsonResult, readNumberParam, readStringParam } from "openclaw/plugin-sdk";
import { fetchPendingTokens, submitAnalysis } from "./server-api.js";

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
      const result = await submitAnalysis(tokenAddress, analysis as any, options?.serverUrl);
      return jsonResult({ ok: true, result });
    }
  };
}
