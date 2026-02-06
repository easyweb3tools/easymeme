import crypto from "crypto";
import { AnalyzeTokenRiskOutput, PendingToken, RiskBand, RiskLevel } from "../types";
import { requireEnv } from "./env";

const DEFAULT_MODEL = "gemini-1.5-pro";

function normalizeBand(input: string | undefined): RiskBand {
  const upper = (input ?? "").toUpperCase();
  if (upper === "LOW" || upper === "MEDIUM" || upper === "HIGH") {
    return upper;
  }
  return "MEDIUM";
}

function normalizeLevel(input: string | undefined): RiskLevel {
  const upper = (input ?? "").toUpperCase();
  if (upper === "SAFE" || upper === "WARNING" || upper === "DANGER") {
    return upper;
  }
  return "WARNING";
}

function safeJsonParse(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return null;
  }
}

function buildFallbackAnalysis(): AnalyzeTokenRiskOutput {
  return {
    riskScore: 50,
    riskLevel: "WARNING",
    isGoldenDog: false,
    riskFactors: {
      honeypotRisk: "MEDIUM",
      taxRisk: "MEDIUM",
      ownerRisk: "MEDIUM",
      concentrationRisk: "MEDIUM"
    },
    reasoning: "LLM unavailable; using fallback risk profile.",
    recommendation: "请谨慎观察，等待更多链上数据确认后再操作。"
  };
}

async function callGemini(prompt: string): Promise<string> {
  const apiKey = requireEnv("GEMINI_API_KEY");
  const model = process.env.GEMINI_MODEL?.trim() || DEFAULT_MODEL;
  const url = `https://generativelanguage.googleapis.com/v1beta/models/${model}:generateContent?key=${apiKey}`;

  const body = {
    contents: [{ role: "user", parts: [{ text: prompt }] }],
    generationConfig: {
      temperature: 0.2,
      responseMimeType: "application/json"
    }
  };

  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body)
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Gemini API error: ${res.status} ${text}`);
  }

  const payload = (await res.json()) as {
    candidates?: Array<{ content?: { parts?: Array<{ text?: string }> } }>;
  };
  const text = payload.candidates?.[0]?.content?.parts?.[0]?.text;
  if (!text) {
    throw new Error("Gemini API returned empty response.");
  }
  return text;
}

function buildPrompt(token: PendingToken): string {
  return [
    "你是 BNB Chain 上的 Meme 币风险分析专家。",
    "请基于给定代币信息输出严格 JSON（无多余文字），格式如下：",
    "{",
    '  "riskScore": 0-100,',
    '  "riskLevel": "SAFE|WARNING|DANGER",',
    '  "isGoldenDog": true|false,',
    '  "riskFactors": {',
    '    "honeypotRisk": "LOW|MEDIUM|HIGH",',
    '    "taxRisk": "LOW|MEDIUM|HIGH",',
    '    "ownerRisk": "LOW|MEDIUM|HIGH",',
    '    "concentrationRisk": "LOW|MEDIUM|HIGH"',
    "  },",
    '  "reasoning": "简短原因",',
    '  "recommendation": "给用户的建议"',
    "}",
    "",
    "代币信息如下：",
    JSON.stringify(token, null, 2)
  ].join("\n");
}

export async function analyzeRisk(token: PendingToken): Promise<{
  analysis: AnalyzeTokenRiskOutput;
  patternHash: string;
  patternDescription: string;
}> {
  let analysis = buildFallbackAnalysis();

  try {
    const prompt = buildPrompt(token);
    const responseText = await callGemini(prompt);
    const parsed = safeJsonParse(responseText);
    if (parsed && typeof parsed === "object") {
      const obj = parsed as Partial<AnalyzeTokenRiskOutput>;
      analysis = {
        riskScore: typeof obj.riskScore === "number" ? obj.riskScore : 50,
        riskLevel: normalizeLevel(obj.riskLevel),
        isGoldenDog: Boolean(obj.isGoldenDog),
        riskFactors: {
          honeypotRisk: normalizeBand(obj.riskFactors?.honeypotRisk),
          taxRisk: normalizeBand(obj.riskFactors?.taxRisk),
          ownerRisk: normalizeBand(obj.riskFactors?.ownerRisk),
          concentrationRisk: normalizeBand(obj.riskFactors?.concentrationRisk)
        },
        reasoning: typeof obj.reasoning === "string" ? obj.reasoning : "无明确说明",
        recommendation: typeof obj.recommendation === "string" ? obj.recommendation : "请谨慎评估风险。"
      };
    }
  } catch (error) {
    analysis = {
      ...analysis,
      reasoning: `LLM 调用失败: ${(error as Error).message}`
    };
  }

  const patternRaw = JSON.stringify({
    riskLevel: analysis.riskLevel,
    riskFactors: analysis.riskFactors,
    isGoldenDog: analysis.isGoldenDog
  });
  const patternHash = crypto.createHash("sha256").update(patternRaw).digest("hex");
  const patternDescription = `${analysis.riskLevel} ${analysis.riskFactors.honeypotRisk}/${analysis.riskFactors.taxRisk}/${analysis.riskFactors.ownerRisk}/${analysis.riskFactors.concentrationRisk}`;

  return { analysis, patternHash, patternDescription };
}
