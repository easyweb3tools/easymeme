import { fetchPendingTokens as fetchFromServer } from "../services/serverApi";
import { loadTokenHistory, saveTokenHistory, ToolContext } from "../services/memory";
import { FetchPendingTokensInput, PendingToken } from "../types";

export async function fetchPendingTokens(
  input: FetchPendingTokensInput = {},
  ctx?: ToolContext
): Promise<PendingToken[]> {
  const limit = input.limit ?? 10;
  const history = await loadTokenHistory(ctx);
  const candidates = await fetchFromServer(limit);

  const results: PendingToken[] = [];
  for (const token of candidates) {
    const tokenAddress = token.address.toLowerCase();
    if (history[tokenAddress]) {
      continue;
    }
    results.push(token);
    history[tokenAddress] = {
      tokenAddress,
      firstSeenAt: token.createdAt ?? new Date().toISOString(),
      lastAnalyzedAt: "",
      analysisHistory: [],
      userInterest: false
    };
  }

  await saveTokenHistory(ctx, history);
  return results;
}
