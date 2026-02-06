import { analyzeRisk } from "../services/riskAnalyzer";
import {
  loadGoldenDogHistory,
  loadRiskPatterns,
  loadTokenHistory,
  saveGoldenDogHistory,
  saveRiskPatterns,
  saveTokenHistory,
  ToolContext
} from "../services/memory";
import { AnalyzeTokenRiskInput, AnalyzeTokenRiskOutput } from "../types";

export async function analyzeTokenRisk(
  input: AnalyzeTokenRiskInput,
  ctx?: ToolContext
): Promise<AnalyzeTokenRiskOutput> {
  const tokenAddress = input.token.address.toLowerCase();
  const { analysis, patternHash, patternDescription } = await analyzeRisk(input.token);

  const history = await loadTokenHistory(ctx);
  const existing = history[tokenAddress];
  const now = new Date().toISOString();
  history[tokenAddress] = {
    tokenAddress,
    firstSeenAt: existing?.firstSeenAt ?? now,
    lastAnalyzedAt: now,
    analysisHistory: [...(existing?.analysisHistory ?? []), analysis],
    userInterest: existing?.userInterest ?? false
  };
  await saveTokenHistory(ctx, history);

  const patterns = await loadRiskPatterns(ctx);
  const existingPattern = patterns[patternHash];
  patterns[patternHash] = {
    patternHash,
    description: patternDescription,
    count: (existingPattern?.count ?? 0) + 1
  };
  await saveRiskPatterns(ctx, patterns);

  if (analysis.isGoldenDog) {
    const goldenDogs = await loadGoldenDogHistory(ctx);
    goldenDogs[tokenAddress] = {
      tokenAddress,
      detectedAt: now,
      reasoning: analysis.reasoning
    };
    await saveGoldenDogHistory(ctx, goldenDogs);
  }

  return analysis;
}
