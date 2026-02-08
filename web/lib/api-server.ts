import { headers } from 'next/headers';
import type { AITrade, AITradeStats, GoldenDogToken, Token, TokenDetail } from './api-types';

function getServerBaseUrl(): string {
  const h = headers();
  const proto = h.get('x-forwarded-proto') || 'http';
  const host = h.get('x-forwarded-host') || h.get('host');
  if (!host) {
    return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  }
  return `${proto}://${host}`;
}

export async function getGoldenDogs(limit = 50): Promise<GoldenDogToken[]> {
  const base = getServerBaseUrl();
  const res = await fetch(`${base}/api/tokens/golden-dogs?limit=${limit}`, {
    cache: 'no-store',
  });
  const data = await res.json();
  return data.data;
}

export async function getTokenDetail(address: string): Promise<TokenDetail> {
  const base = getServerBaseUrl();
  const res = await fetch(`${base}/api/tokens/${address}/detail`, {
    cache: 'no-store',
  });
  const data = await res.json();
  return data.data;
}

export async function getTokens(): Promise<Token[]> {
  const base = getServerBaseUrl();
  const res = await fetch(`${base}/api/tokens`, { cache: 'no-store' });
  const data = await res.json();
  return data.data;
}

export async function getToken(address: string): Promise<Token> {
  const base = getServerBaseUrl();
  const res = await fetch(`${base}/api/tokens/${address}`, { cache: 'no-store' });
  const data = await res.json();
  return data.data;
}

export async function getAITrades(limit = 50): Promise<AITrade[]> {
  const base = getServerBaseUrl();
  const res = await fetch(`${base}/api/ai-trades?limit=${limit}`, {
    cache: 'no-store',
  });
  const data = await res.json();
  return data.data;
}

export async function getAITradeStats(): Promise<AITradeStats> {
  const base = getServerBaseUrl();
  const res = await fetch(`${base}/api/ai-trades/stats`, {
    cache: 'no-store',
  });
  const data = await res.json();
  return data.data;
}
