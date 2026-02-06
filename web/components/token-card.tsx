'use client';

import { Token } from '@/lib/api';
import { RiskBadge } from './risk-badge';
import { TradePanel } from './trade-panel';
import { useState } from 'react';

export function TokenCard({ token }: { token: Token }) {
  const [showTrade, setShowTrade] = useState(false);

  return (
    <div className="border rounded-lg p-4 bg-card">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <RiskBadge level={token.risk_level} score={token.risk_score} />
          <div>
            <h3 className="font-bold">${token.symbol || 'Unknown'}</h3>
            <p className="text-sm text-muted-foreground truncate w-40">
              {token.address}
            </p>
          </div>
        </div>

        <div className="text-right">
          <p className="text-sm">
            LP: {parseFloat(token.initial_liquidity).toFixed(2)} BNB
          </p>
          <p className="text-sm text-muted-foreground">
            Tax: {token.buy_tax}% / {token.sell_tax}%
          </p>
        </div>

        <div className="flex gap-2">
          <a
            href={`https://bscscan.com/token/${token.address}`}
            target="_blank"
            rel="noopener noreferrer"
            className="px-3 py-1 border rounded text-sm hover:bg-accent"
          >
            View
          </a>
          {!token.is_honeypot &&
            (token.risk_level === 'safe' || token.risk_level === 'warning') && (
            <button
              onClick={() => setShowTrade(!showTrade)}
              className="px-3 py-1 bg-primary text-primary-foreground rounded text-sm"
            >
              Buy
            </button>
          )}
        </div>
      </div>

      {showTrade && (
        <div className="mt-4 pt-4 border-t">
          <TradePanel token={token} />
        </div>
      )}
    </div>
  );
}
