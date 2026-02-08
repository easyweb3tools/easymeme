import type { AITrade, AITradeStats, GoldenDogToken, Token, TokenDetail } from './api-types';

export async function getGoldenDogs(limit = 50): Promise<GoldenDogToken[]> {
  const res = await fetch(`/api/tokens/golden-dogs?limit=${limit}`, {
    cache: 'no-store',
  });
  const data = await res.json();
  return data.data;
}

export async function getTokenDetail(address: string): Promise<TokenDetail> {
  const res = await fetch(`/api/tokens/${address}/detail`, {
    cache: 'no-store',
  });
  const data = await res.json();
  return data.data;
}

export async function getTokens(): Promise<Token[]> {
  const res = await fetch(`/api/tokens`, { cache: 'no-store' });
  const data = await res.json();
  return data.data;
}

export async function getToken(address: string): Promise<Token> {
  const res = await fetch(`/api/tokens/${address}`, { cache: 'no-store' });
  const data = await res.json();
  return data.data;
}

export function createWebSocket(onMessage: (data: any) => void): WebSocket {
  const wsUrl =
    process.env.NEXT_PUBLIC_WS_URL ||
    `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`;
  const ws = new WebSocket(wsUrl);

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    onMessage(data);
  };

  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };

  return ws;
}

export async function getAITrades(limit = 50): Promise<AITrade[]> {
  const res = await fetch(`/api/ai-trades?limit=${limit}`, {
    cache: 'no-store',
  });
  const data = await res.json();
  return data.data;
}

export async function getAITradeStats(): Promise<AITradeStats> {
  const res = await fetch(`/api/ai-trades/stats`, {
    cache: 'no-store',
  });
  const data = await res.json();
  return data.data;
}
