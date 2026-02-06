'use client';

import { useEffect, useState } from 'react';
import { Token, getTokens, createWebSocket } from '@/lib/api';
import { TokenCard } from './token-card';

export function TokenList() {
  const [tokens, setTokens] = useState<Token[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let alive = true;

    const load = async () => {
      try {
        const data = await getTokens();
        if (alive) {
          setTokens(data);
          setLoading(false);
        }
      } catch (err) {
        if (alive) {
          setLoading(false);
        }
      }
    };

    load();

    const interval = setInterval(load, 10000);

    const ws = createWebSocket((data) => {
      if (data.type === 'new_token') {
        setTokens((prev) => [data.token, ...prev].slice(0, 50));
      }
    });

    return () => {
      alive = false;
      clearInterval(interval);
      ws.close();
    };
  }, []);

  if (loading) {
    return <div className="text-center py-8">Loading...</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">New Tokens</h2>
        <span className="text-sm text-green-500 animate-pulse">● Live</span>
      </div>
      <div className="grid gap-4">
        {tokens.map((token) => (
          <TokenCard key={token.id} token={token} />
        ))}
      </div>
    </div>
  );
}
