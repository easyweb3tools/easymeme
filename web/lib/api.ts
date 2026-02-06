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

export async function getTokens(): Promise<Token[]> {
  const res = await fetch(`${API_URL}/api/tokens`);
  const data = await res.json();
  return data.data;
}

export async function getToken(address: string): Promise<Token> {
  const res = await fetch(`${API_URL}/api/tokens/${address}`);
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
