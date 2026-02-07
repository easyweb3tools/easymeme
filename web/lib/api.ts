const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws';

export interface Token {
  id: string;
  address: string;
  name: string;
  symbol: string;
  decimals: number;
  pair_address: string;
  initial_liquidity: string;
  risk_score: number;
  risk_level: 'pending' | 'safe' | 'warning' | 'danger';
  is_honeypot: boolean;
  buy_tax: number;
  sell_tax: number;
  created_at: string;
}
export type GoldenDogToken = {
  address: string;
  name: string;
  symbol: string;
  pairAddress: string;
  liquidity: number;
  riskScore: number;
  riskLevel: 'pending' | 'safe' | 'warning' | 'danger';
  isGoldenDog: boolean;
  goldenDogScore: number;
  effectiveScore: number;
  timeDecayFactor: number;
  phase: 'EARLY' | 'PEAK' | 'DECLINING' | 'EXPIRED';
  createdAt: string;
  analyzedAt?: string | null;
};

export type TokenDetail = {
  id: string;
  address: string;
  name: string;
  symbol: string;
  pairAddress: string;
  dex: string;
  liquidity: number;
  creatorAddress: string;
  createdAt: string;
  analyzedAt?: string | null;
  riskScore: number;
  riskLevel: 'pending' | 'safe' | 'warning' | 'danger';
  isGoldenDog: boolean;
  goldenDogScore: number;
  effectiveScore: number;
  timeDecayFactor: number;
  phase: 'EARLY' | 'PEAK' | 'DECLINING' | 'EXPIRED';
  riskDetails?: Record<string, unknown>;
  analysisResult?: Record<string, unknown>;
};

export type AITrade = {
  id: string;
  token_address: string;
  token_symbol: string;
  type: 'BUY' | 'SELL';
  amount_in: string;
  amount_out: string;
  tx_hash: string;
  timestamp: string;
  status: string;
  gas_used: string;
  block_number: number;
  error_message: string;
  golden_dog_score: number;
  decision_reason: string;
  strategy_used: string;
  current_value: string;
  profit_loss: number;
  user_id: string;
};

export type AITradeStats = {
  count: number;
  winRate: number;
  avgPL: number;
  totalPL: number;
  byStrategy: Array<{
    strategy: string;
    count: number;
    winRate: number;
    avgPL: number;
    totalPL: number;
  }>;
  byPeriod: Array<{
    period: string;
    count: number;
    winRate: number;
    avgPL: number;
    totalPL: number;
  }>;
};

export async function getGoldenDogs(limit = 50): Promise<GoldenDogToken[]> {
  const res = await fetch(`${API_URL}/api/tokens/golden-dogs?limit=${limit}`, {
    cache: 'no-store',
  });
  const data = await res.json();
  return data.data;
}

export async function getTokenDetail(address: string): Promise<TokenDetail> {
  const res = await fetch(`${API_URL}/api/tokens/${address}/detail`, {
    cache: 'no-store',
  });
  const data = await res.json();
  return data.data;
}

export async function getTokens(): Promise<Token[]> {
  const res = await fetch(`${API_URL}/api/tokens`, { cache: 'no-store' });
  const data = await res.json();
  return data.data;
}

export async function getToken(address: string): Promise<Token> {
  const res = await fetch(`${API_URL}/api/tokens/${address}`, { cache: 'no-store' });
  const data = await res.json();
  return data.data;
}

export function createWebSocket(onMessage: (data: any) => void): WebSocket {
  const ws = new WebSocket(WS_URL);

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
  const res = await fetch(`${API_URL}/api/ai-trades?limit=${limit}`, {
    cache: 'no-store',
  });
  const data = await res.json();
  return data.data;
}

export async function getAITradeStats(): Promise<AITradeStats> {
  const res = await fetch(`${API_URL}/api/ai-trades/stats`, {
    cache: 'no-store',
  });
  const data = await res.json();
  return data.data;
}
