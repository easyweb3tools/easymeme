import { Type } from "@sinclair/typebox";
import type { OpenClawPluginApi } from "openclaw/plugin-sdk";
import {
  createAnalyzeTokenRiskTool,
  createBuildAnalysisDraftTool,
  createEstimateGoldenDogScoreTool,
  createFetchPendingTokensTool,
  createExecuteTradeTool,
  createGetPositionsTool,
  createGetRulePerformanceReportTool,
  createGetWalletInfoTool,
  createRecordOutcomeTool,
  createRecordUserFeedbackTool,
  createSubmitAnalysisTool,
  createUpsertWalletConfigTool
} from "./src/tools.js";

const plugin = {
  id: "easymeme-openclaw-skill",
  name: "EasyMeme OpenClaw Skill",
  description: "EasyMeme analysis tools for BNB Chain tokens.",
  configSchema: Type.Object(
    {
      serverUrl: Type.Optional(Type.String()),
      userId: Type.Optional(Type.String())
    },
    { additionalProperties: false }
  ),
  register(api: OpenClawPluginApi) {
    api.registerTool((ctx) =>
      createFetchPendingTokensTool({ serverUrl: (ctx.config as any)?.serverUrl })
    );
    api.registerTool(createBuildAnalysisDraftTool());
    api.registerTool(createAnalyzeTokenRiskTool());
    api.registerTool(createEstimateGoldenDogScoreTool());
    api.registerTool((ctx) =>
      createSubmitAnalysisTool({ serverUrl: (ctx.config as any)?.serverUrl })
    );
    api.registerTool((ctx) =>
      createExecuteTradeTool({
        serverUrl: (ctx.config as any)?.serverUrl,
        userId: (ctx.config as any)?.userId
      })
    );
    api.registerTool((ctx) =>
      createUpsertWalletConfigTool({
        serverUrl: (ctx.config as any)?.serverUrl,
        userId: (ctx.config as any)?.userId
      })
    );
    api.registerTool((ctx) =>
      createGetWalletInfoTool({
        serverUrl: (ctx.config as any)?.serverUrl,
        userId: (ctx.config as any)?.userId
      })
    );
    api.registerTool((ctx) =>
      createGetPositionsTool({
        serverUrl: (ctx.config as any)?.serverUrl,
        userId: (ctx.config as any)?.userId
      })
    );
    api.registerTool(createGetRulePerformanceReportTool());
    api.registerTool(createRecordOutcomeTool());
    api.registerTool(createRecordUserFeedbackTool());
  }
};

export default plugin;
