'use client';

import { useState } from 'react';

export function CopyButton({
  value,
  label = 'Copy',
  copiedLabel = 'Copied',
}: {
  value: string;
  label?: string;
  copiedLabel?: string;
}) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      setCopied(false);
    }
  };

  return (
    <button
      type="button"
      onClick={handleCopy}
      className="text-xs text-white/60 hover:text-white border border-white/20 rounded-full px-3 py-1"
    >
      {copied ? copiedLabel : label}
    </button>
  );
}
