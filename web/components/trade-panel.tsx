'use client';

import { useState } from 'react';
import { useAccount, usePublicClient, useWriteContract, useWaitForTransactionReceipt } from 'wagmi';
import { parseEther } from 'viem';
import { Token } from '@/lib/api';

const PANCAKE_ROUTER = process.env.NEXT_PUBLIC_PANCAKE_ROUTER || '';
const WBNB = process.env.NEXT_PUBLIC_WBNB || '';

const ROUTER_ABI = [
  {
    name: 'swapExactETHForTokensSupportingFeeOnTransferTokens',
    type: 'function',
    inputs: [
      { name: 'amountOutMin', type: 'uint256' },
      { name: 'path', type: 'address[]' },
      { name: 'to', type: 'address' },
      { name: 'deadline', type: 'uint256' },
    ],
    outputs: [],
    stateMutability: 'payable',
  },
  {
    name: 'getAmountsOut',
    type: 'function',
    inputs: [
      { name: 'amountIn', type: 'uint256' },
      { name: 'path', type: 'address[]' },
    ],
    outputs: [{ name: 'amounts', type: 'uint256[]' }],
    stateMutability: 'view',
  },
] as const;

const AMOUNTS = [0.1, 0.5, 1, 5];

export function TradePanel({ token }: { token: Token }) {
  const [amount, setAmount] = useState(0.1);
  const [slippage, setSlippage] = useState(0.5);
  const { address, isConnected } = useAccount();
  const publicClient = usePublicClient();

  const { writeContract, data: hash, isPending } = useWriteContract();

  const { isLoading: isConfirming, isSuccess } = useWaitForTransactionReceipt({
    hash,
  });

  const handleBuy = async () => {
    if (!address) return;
    if (!PANCAKE_ROUTER || !WBNB) return;

    const deadline = BigInt(Math.floor(Date.now() / 1000) + 1200);
    let amountOutMin = 0n;
    if (publicClient) {
      try {
        const amountIn = parseEther(amount.toString());
        const amounts = (await publicClient.readContract({
          address: PANCAKE_ROUTER as `0x${string}`,
          abi: ROUTER_ABI,
          functionName: 'getAmountsOut',
          args: [amountIn, [WBNB as `0x${string}`, token.address as `0x${string}`]],
        })) as bigint[];
        const expectedOut = amounts[amounts.length - 1] ?? 0n;
        const slippageBps = BigInt(Math.round(slippage * 100));
        amountOutMin = (expectedOut * (10000n - slippageBps)) / 10000n;
      } catch {
        amountOutMin = 0n;
      }
    }

    writeContract({
      address: PANCAKE_ROUTER as `0x${string}`,
      abi: ROUTER_ABI,
      functionName: 'swapExactETHForTokensSupportingFeeOnTransferTokens',
      args: [amountOutMin, [WBNB as `0x${string}`, token.address as `0x${string}`], address, deadline],
      value: parseEther(amount.toString()),
    });
  };

  if (!isConnected) {
    return <p className="text-sm text-muted-foreground">Connect wallet to trade</p>;
  }
  if (!PANCAKE_ROUTER || !WBNB) {
    return (
      <p className="text-sm text-muted-foreground">
        Missing router configuration. Set NEXT_PUBLIC_PANCAKE_ROUTER and NEXT_PUBLIC_WBNB.
      </p>
    );
  }

  return (
    <div className="space-y-4">
      <div>
        <p className="text-sm mb-2">Amount (BNB)</p>
        <div className="flex gap-2">
          {AMOUNTS.map((a) => (
            <button
              key={a}
              onClick={() => setAmount(a)}
              className={`px-4 py-2 rounded border ${
                amount === a ? 'bg-primary text-primary-foreground' : ''
              }`}
            >
              {a}
            </button>
          ))}
        </div>
      </div>
      <div>
        <label className="text-sm mb-2 block">Slippage (%)</label>
        <input
          type="number"
          min={0.1}
          max={5}
          step={0.1}
          value={slippage}
          onChange={(e) => setSlippage(Number(e.target.value))}
          className="w-full rounded border px-3 py-2 text-sm"
        />
      </div>

      <button
        onClick={handleBuy}
        disabled={isPending || isConfirming}
        className="w-full py-3 bg-green-500 text-white rounded font-bold disabled:opacity-50"
      >
        {isPending
          ? 'Confirming...'
          : isConfirming
          ? 'Processing...'
          : `Buy with ${amount} BNB`}
      </button>

      {isSuccess && (
        <p className="text-green-500 text-sm">
          Transaction successful!{' '}
          <a
            href={`https://bscscan.com/tx/${hash}`}
            target="_blank"
            rel="noopener noreferrer"
            className="underline"
          >
            View on BSCScan
          </a>
        </p>
      )}
    </div>
  );
}
