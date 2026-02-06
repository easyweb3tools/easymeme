import { fetchPendingTokens } from "./tools/fetchPendingTokens";
import { analyzeTokenRisk } from "./tools/analyzeTokenRisk";
import { submitAnalysis } from "./tools/submitAnalysis";

type ToolExecute = (toolCallId: string, params: Record<string, unknown>) => Promise<unknown>;

const fetchPendingTokensTool = {
  name: "fetchPendingTokens",
  label: "Fetch Pending Tokens",
  description: "Fetch pending tokens from the server for AI analysis.",
  parameters: {
    type: "object",
    properties: {
      limit: { type: "number" }
    },
    additionalProperties: false
  },
  execute: (async (_toolCallId, params) => {
    const limit = typeof params.limit === "number" ? params.limit : undefined;
    return fetchPendingTokens({ limit });
  }) as ToolExecute
};

const analyzeTokenRiskTool = {
  name: "analyzeTokenRisk",
  label: "Analyze Token Risk",
  description: "Use LLM to assess token risk and determine if it is a golden dog.",
  parameters: {
    type: "object",
    properties: {
      token: {
        type: "object",
        properties: {
          address: { type: "string" },
          name: { type: "string" },
          symbol: { type: "string" },
          liquidity: { type: "number" },
          creatorAddress: { type: "string" },
          contractCode: { type: "string" },
          holderDistribution: { type: "array" },
          creatorHistory: { type: "array" }
        },
        required: ["address", "name", "symbol", "liquidity", "creatorAddress"],
        additionalProperties: true
      }
    },
    required: ["token"],
    additionalProperties: false
  },
  execute: (async (_toolCallId, params) => {
    const token = typeof params.token === "object" && params.token ? (params.token as Record<string, unknown>) : null;
    if (!token) {
      throw new Error("token is required");
    }
    return analyzeTokenRisk({
      token: {
        address: String(token.address ?? ""),
        name: String(token.name ?? ""),
        symbol: String(token.symbol ?? ""),
        liquidity: Number(token.liquidity ?? 0),
        creatorAddress: String(token.creatorAddress ?? ""),
        contractCode: typeof token.contractCode === "string" ? token.contractCode : undefined,
        holderDistribution: Array.isArray(token.holderDistribution) ? (token.holderDistribution as any) : undefined,
        creatorHistory: Array.isArray(token.creatorHistory) ? (token.creatorHistory as any) : undefined,
        createdAt: typeof token.createdAt === "string" ? token.createdAt : undefined,
        pairAddress: typeof token.pairAddress === "string" ? token.pairAddress : undefined
      }
    });
  }) as ToolExecute
};

const submitAnalysisTool = {
  name: "submitAnalysis",
  label: "Submit Analysis",
  description: "Submit the AI analysis result back to the server.",
  parameters: {
    type: "object",
    properties: {
      tokenAddress: { type: "string" },
      analysis: { type: "object" }
    },
    required: ["tokenAddress", "analysis"],
    additionalProperties: false
  },
  execute: (async (_toolCallId, params) => {
    const tokenAddress = typeof params.tokenAddress === "string" ? params.tokenAddress : "";
    if (!tokenAddress) {
      throw new Error("tokenAddress is required");
    }
    const analysis = typeof params.analysis === "object" && params.analysis ? (params.analysis as any) : null;
    if (!analysis) {
      throw new Error("analysis is required");
    }
    return submitAnalysis({ tokenAddress, analysis });
  }) as ToolExecute
};

const plugin = {
  id: "easymeme-openclaw-skill",
  name: "EasyMeme OpenClaw Skill",
  version: "0.1.0",
  description: "OpenClaw plugin that exposes EasyMeme analysis tools.",
  register: (api: any) => {
    api.registerTool(fetchPendingTokensTool);
    api.registerTool(analyzeTokenRiskTool);
    api.registerTool(submitAnalysisTool);
  }
};

export default plugin;
