type NotifyPayload = {
  tokenAddress: string;
  tokenSymbol?: string;
  amountIn?: string;
  txHash?: string;
};

function resolveTelegramToken(): string | undefined {
  return (
    process.env.EASYMEME_NOTIFY_TOKEN?.trim() ||
    process.env.TELEGRAM_BOT_TOKEN?.trim() ||
    process.env.OPENCLAW_TELEGRAM_BOT_TOKEN?.trim()
  );
}

function extractTxHash(result: unknown): string | undefined {
  if (!result || typeof result !== "object") {
    return undefined;
  }
  const record = result as Record<string, unknown>;
  const data = record.data as Record<string, unknown> | undefined;
  const hash = (data?.tx_hash ?? record.tx_hash) as string | undefined;
  return typeof hash === "string" && hash.length > 0 ? hash : undefined;
}

export async function notifySellTrade(
  payload: NotifyPayload,
  result: unknown
): Promise<void> {
  const channel = process.env.EASYMEME_NOTIFY_CHANNEL?.trim().toLowerCase();
  if (channel !== "telegram" && channel !== "tg") {
    return;
  }
  const to = process.env.EASYMEME_NOTIFY_TO?.trim();
  if (!to) {
    return;
  }
  const token = resolveTelegramToken();
  if (!token) {
    return;
  }

  const txHash = payload.txHash || extractTxHash(result);
  const lines = [
    "EasyMeme SELL executed",
    payload.tokenSymbol ? `Token: ${payload.tokenSymbol}` : "Token: (unknown)",
    `Address: ${payload.tokenAddress}`,
    payload.amountIn ? `AmountIn: ${payload.amountIn}` : "AmountIn: (unknown)",
  ];
  if (txHash) {
    lines.push(`Tx: ${txHash}`);
    lines.push(`BscScan: https://bscscan.com/tx/${txHash}`);
  }

  const text = lines.join("\n");
  const url = `https://api.telegram.org/bot${token}/sendMessage`;
  try {
    await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        chat_id: to,
        text,
        disable_web_page_preview: true
      })
    });
  } catch (err) {
    console.warn("telegram notify failed", err);
  }
}
