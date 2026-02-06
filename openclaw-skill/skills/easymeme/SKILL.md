---
name: easymeme
description: Autonomous EasyMeme agent that analyzes pending BNB Chain tokens and writes risk analysis back to the EasyMeme server using OpenClaw tools.
---

# EasyMeme Agent

## Purpose

Continuously analyze pending BNB Chain tokens from the EasyMeme server, decide if each token is a "golden dog", and write structured analysis back to the server. Use the OpenClaw model/runtime for analysis; do not call external AI APIs directly.

## Required tools

- `fetchPendingTokens`
- `analyzeTokenRisk`
- `submitAnalysis`

## Workflow (follow in order)

1. Call `fetchPendingTokens` (default limit 10). If the list is empty, reply that there are no pending tokens and stop.
2. For each token:
   - Produce an analysis JSON (as plain text first) in the exact schema required by `analyzeTokenRisk`.
   - Verify the JSON includes all required fields.
   - Then call `analyzeTokenRisk` with `{ token, analysis }`.
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
- `riskFactors`: honeypotRisk, taxRisk, ownerRisk, concentrationRisk (LOW | MEDIUM | HIGH)
- `reasoning`: concise explanation referencing observed data
- `recommendation`: short user-facing suggestion

Before calling `submitAnalysis`, construct a complete JSON object that includes every field above. Do not omit any field.

Minimal example (structure only; values must be your analysis):

```json
{
  "riskScore": 42,
  "riskLevel": "WARNING",
  "isGoldenDog": false,
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

## Error handling

- If the server returns invalid data, skip that token and continue.
- If submission fails, report the failure in the final summary.

## Environment

- Set `EASYMEME_SERVER_URL` to target the EasyMeme server (e.g. `http://server:8080`).
