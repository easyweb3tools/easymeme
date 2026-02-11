---
name: easymeme
description: Autonomous EasyMeme agent that analyzes pending BNB Chain tokens and writes risk analysis back to the EasyMeme server using OpenClaw tools.
---

# EasyMeme Agent

## Purpose

Continuously analyze pending BNB Chain tokens from the EasyMeme server, decide if each token is a "golden dog", and write structured analysis back to the server. Use the OpenClaw model/runtime for analysis; do not call external AI APIs directly.

## Required tools

- `fetchPendingTokens`
- `buildAnalysisDraft`
- `analyzeTokenRisk`
- `estimateGoldenDogScore`
- `submitAnalysis`
- `executeTrade`
- `upsertWalletConfig`
- `recordOutcome`
- `getRulePerformanceReport`
- `recordUserFeedback`

## Workflow (follow in order)

1. Call `fetchPendingTokens` (default limit 10). If the list is empty, reply that there are no pending tokens and stop.
2. For each token:
   - First call `buildAnalysisDraft` to generate a deterministic baseline analysis from enriched fields.
   - Then refine the analysis JSON (as plain text first) in the exact schema required by `analyzeTokenRisk`.
   - Prefer `token.goplus` and `token.dexscreener` fields as primary evidence (do not invent missing data).
   - Call `estimateGoldenDogScore` using the analysis inputs to get a learned `goldenDogScore`.
   - Verify the JSON includes all required fields.
   - Then call `analyzeTokenRisk` with `{ token, analysis }`.
   - If `isGoldenDog` and score >= threshold (default 75), call `executeTrade`.
   - For SELL, include `profitLoss` to allow stop-loss/take-profit gating.
   - Call `submitAnalysis` with `{ tokenAddress, analysis }` to persist the result.
3. Summarize how many tokens were analyzed and submitted.

## Analysis requirements

Your analysis must include ALL required fields. If any field is missing, the tool call will fail.

Two-step rule (critical):
1) First, construct the full JSON object as plain text and visually verify every required field exists.
2) Then pass that JSON object into `analyzeTokenRisk` and `submitAnalysis`.

- `riskScore`: 0-100
- `riskLevel`: SAFE | WARNING | DANGER
- `isGoldenDog`: true only if the token is worth close monitoring, not just "safe"
- `goldenDogScore`: 0-100 (auto-filled if omitted; still preferred to set explicitly)
- `riskFactors`: honeypotRisk, taxRisk, ownerRisk, concentrationRisk (LOW | MEDIUM | HIGH)
- `reasoning`: concise explanation referencing observed data
- `recommendation`: short user-facing suggestion

Risk mapping baseline:
- `token.goplus.is_honeypot = "1"` -> `honeypotRisk: HIGH`
- High `buy_tax` / `sell_tax` from GoPlus -> raise `taxRisk`
- GoPlus owner privileges (e.g. mint/proxy/take ownership) -> raise `ownerRisk`
- DEXScreener txns/liquidity imbalance + holder concentration clues -> raise `concentrationRisk`

Before calling `submitAnalysis`, construct a complete JSON object that includes every field above. Do not omit any field.

Minimal example (structure only; values must be your analysis):

```json
{
  "riskScore": 42,
  "riskLevel": "WARNING",
  "isGoldenDog": false,
  "goldenDogScore": 35,
  "riskFactors": {
    "honeypotRisk": "LOW",
    "taxRisk": "MEDIUM",
    "ownerRisk": "LOW",
    "concentrationRisk": "HIGH"
  },
  "reasoning": "Short reasoning based on available data.",
  "recommendation": "Short user-facing suggestion."
}
```

## Golden dog definition

A "golden dog" is a token with credible upside potential. It is not necessarily safe. You must consider:

- Security (non-honeypot, reasonable tax, no dangerous privileges)
- Liquidity (sufficient and ideally locked)
- Holder distribution (not too concentrated)
- Creator history (no rug history)
- Community interest (if data available)

If the token is safe but low potential, mark `isGoldenDog = false`.

## Golden dog score estimation (temporary rule-based)

`estimateGoldenDogScore` reads OpenClaw local memory weights and returns a learned score.
If the tool fails, fall back to the rule/weight mix below:

1. Start with `riskScore` as base.
2. If `isGoldenDog = true`, add +15; if false, subtract -10.
3. Adjust for risk factors:
   - Any `HIGH` factor: -15 each
   - Any `MEDIUM` factor: -5 each
4. Clamp to 0-100.

Quick formula example:

```
goldenDogScore =
  clamp(0, 100,
    riskScore
    + (isGoldenDog ? 15 : -10)
    - 15 * count(HIGH)
    - 5 * count(MEDIUM)
  )
```

Use this estimate consistently for now, and keep `reasoning` aligned with the adjustments.

## Error handling

- If the server returns invalid data, skip that token and continue.
- If submission fails, report the failure in the final summary.

## Environment

- Set `EASYMEME_SERVER_URL` to target the EasyMeme server (e.g. `http://server:8080`).
- Optional: set `EASYMEME_MEMORY_PATH` to store learned weights (default `~/.easymeme/memory.json`).
- Optional: set `EASYMEME_USER_ID` to bind managed wallet and AI trades (default `default`).

## Learning trigger

When a trade outcome is known, call `recordOutcome` with the result to update local memory weights.

## Example learning loop (concise)

1. Estimate score before trade:
```json
{
  "riskScore": 78,
  "isGoldenDog": true,
  "riskFactors": {
    "honeypotRisk": "LOW",
    "taxRisk": "MEDIUM",
    "ownerRisk": "LOW",
    "concentrationRisk": "MEDIUM"
  }
}
```

2. Record outcome after trade:
```json
{
  "tokenAddress": "0xabc...",
  "outcome": "MOON",
  "maxGain": 1.6,
  "maxLoss": -0.1,
  "analysis": {
    "isGoldenDog": true,
    "riskFactors": {
      "honeypotRisk": "LOW",
      "taxRisk": "MEDIUM",
      "ownerRisk": "LOW",
      "concentrationRisk": "MEDIUM"
    },
    "confidenceWeight": 0.8
  }
}
```

3. Memory weights update:
- `goldenDogBias` tends to increase
- `highPenalty` / `mediumPenalty` tend to decrease

## End-to-end run flow (with learning trigger)

1. `fetchPendingTokens` -> get pending list  
2. For each token: build analysis JSON  
3. `estimateGoldenDogScore` -> fill `goldenDogScore`  
4. `analyzeTokenRisk` -> validate analysis  
5. `submitAnalysis` -> persist to server  
6. When trade outcome is known: `recordOutcome` -> update weights in local memory

## Auto trade config

Use `upsertWalletConfig` to set auto-trade params per user:
```json
{
  "userId": "default",
  "config": {
    "enabled": true,
    "maxAmountPerTrade": 0.1,
    "minGoldenDogScore": 75,
    "dailyBudget": 1
  }
}
```

## Auto take-profit / stop-loss (server enforcement)

When calling `executeTrade` for SELL, include `profitLoss` (e.g. 0.5 for +50%, -0.3 for -30%).
Server will only allow SELL if:
- `profitLoss <= stopLoss`, or
- `profitLoss >= any takeProfitLevels`, unless `force = true`.

## Telegram feedback -> OpenClaw Memory

When Telegram users send feedback, map it to `recordUserFeedback`:
```json
{
  "tokenAddress": "0xabc...",
  "feedbackType": "REPORT_RUG",
  "userId": "telegram:123456",
  "channel": "TELEGRAM",
  "userReputation": 60
}
```

This will:
- store feedback in local memory
- update weights via `applyFeedback`

## Rule performance tracking

When calling `recordOutcome`, OpenClaw will update local rule performance stats:
- Rule ID: `golden_dog_decision`
- Accuracy is tracked on `MOON` / `RUG` outcomes (ignores `FLAT`)

Tool output now includes:
```json
{
  "rulePerformance": [
    { "ruleId": "golden_dog_decision", "accuracy": 0.72, "correct": 18, "total": 25 }
  ]
}
```
