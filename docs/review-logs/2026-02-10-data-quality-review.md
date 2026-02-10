# EasyMeme Data Quality Review — 2026-02-10

> Reviewer: Web3 Data Scientist Perspective

## Executive Summary

**Current data dimensions are critically insufficient for reliable golden dog identification.** The AI is essentially guessing — it only receives `address + name + symbol + initialLiquidity`, while the fields `creatorAddress`, `buyTax`, `sellTax` in the Token model are **never populated** by the scanner. The `contractCode`, `holderDistribution`, `creatorHistory` fields in the OpenClaw tool schema have **no server-side data source**.

---

## Current Data Coverage

### What `scanner.go` Actually Collects

| Field | Source | Status |
|-------|--------|--------|
| `Address` | PairCreated event Topics | ✅ Available |
| `Name` / `Symbol` / `Decimals` | `GetTokenInfo()` ERC20 call | ✅ Available |
| `PairAddress` | PairCreated event Data | ✅ Available |
| `InitialLiquidity` | `GetPairReserves()` | ✅ Available (snapshot only) |
| `CreatorAddress` | — | ❌ Always empty string |
| `BuyTax` / `SellTax` | — | ❌ Always 0 |
| `IsHoneypot` | — | ❌ Always false |

### What OpenClaw Tool Schema Expects But Server Never Provides

| Field | Expected Source | Status |
|-------|----------------|--------|
| `contractCode` | BSCScan API | ❌ Not implemented |
| `holderDistribution` | BSCScan / on-chain query | ❌ Not implemented |
| `creatorHistory` | BSCScan API | ❌ Not implemented |

---

## Required Data Dimensions for Golden Dog Identification

### P0 — Must Have (Safety / Loss Prevention)

| Dimension | Importance | Recommended Source |
|-----------|------------|-------------------|
| **Honeypot detection** (can you sell?) | CRITICAL | GoPlus API |
| **Real buy/sell tax** | CRITICAL | GoPlus API |
| **Contract open source** | HIGH | GoPlus API |
| **Owner permissions** (mint/pause/blacklist/setFee) | HIGH | GoPlus API |
| **Proxy contract detection** | MEDIUM | GoPlus API |
| **Creator address** | HIGH | PairCreated tx `from` field |

### P1 — Should Have (Market Signal)

| Dimension | Importance | Recommended Source |
|-----------|------------|-------------------|
| **Real-time liquidity changes** | HIGH | Poll Pair Reserves periodically |
| **Price trend** | HIGH | DEXScreener API (free) |
| **Trade count / buy-sell ratio** | HIGH | DEXScreener API |
| **Unique trader count** | MEDIUM | BSCScan / DEXScreener |

### P2 — Nice to Have

| Dimension | Source |
|-----------|--------|
| Creator's historical tokens | BSCScan contract creation records |
| Social media buzz | LunarCrush / Twitter API |
| Smart Money tracking | On-chain whale monitoring |

---

## Recommended Action: GoPlus Security API

**Free, single API call returns 30+ security dimensions:**

```
GET https://api.gopluslabs.io/api/v1/token_security/56?contract_addresses={TOKEN_ADDRESS}
```

Key fields returned:
- `is_honeypot`, `buy_tax`, `sell_tax`
- `is_mintable`, `can_take_back_ownership`, `is_proxy`
- `holder_count`, `lp_holder_count`
- `is_open_source`, `creator_address`, `owner_address`

**Implementation: Scanner should call GoPlus API after detecting a new token, then write results into the `RiskDetails` JSON field.**

---

## Scoring Model Issue

The current `estimateScore()` function:
```
score = riskScore × baseMultiplier + goldenDogBias - penalties
```
All inputs (`riskScore`, `isGoldenDog`, `riskFactors`) come from LLM inference with only 4 data points. The scoring model structure is sound, but the input data quality makes it unreliable.

**Fix: Feed real GoPlus data into the analysis pipeline so the LLM has factual security data to work with.**

---

## Impact Assessment

```
Current:     Data ★☆☆☆☆ | Safety ★☆☆☆☆ | Market ☆☆☆☆☆
+ GoPlus:    Data ★★★★☆ | Safety ★★★★☆ | Market ☆☆☆☆☆
+ DEXScreener: Data ★★★★☆ | Safety ★★★★☆ | Market ★★★☆☆
```
