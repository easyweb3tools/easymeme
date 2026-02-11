type NotifyPayload = {
  tokenAddress: string;
  tokenSymbol?: string;
  amountIn?: string;
  txHash?: string;
};

type GoldenDogNotifyPayload = {
  tokenAddress: string;
  tokenSymbol?: string;
  goldenDogScore?: number;
  riskScore?: number;
  decisionReason?: string;
};

function resolveTelegramToken(): string | undefined {
  return (
    process.env.EASYMEME_NOTIFY_TOKEN?.trim() ||
    process.env.TELEGRAM_BOT_TOKEN?.trim() ||
    process.env.OPENCLAW_TELEGRAM_BOT_TOKEN?.trim()
  );
}

function normalizeTelegramChatId(raw: string): string {
  const value = raw.trim();
  if (value.startsWith("telegram:")) {
    return value.slice("telegram:".length);
  }
  if (value.startsWith("tg:")) {
    return value.slice("tg:".length);
  }
  return value;
}

function isTelegramNotifyEnabled(): boolean {
  const channel = process.env.EASYMEME_NOTIFY_CHANNEL?.trim().toLowerCase();
  return channel === "telegram" || channel === "tg";
}

async function sendTelegramMessage(text: string): Promise<void> {
  if (!isTelegramNotifyEnabled()) {
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

  const chatId = normalizeTelegramChatId(to);
  const url = `https://api.telegram.org/bot${token}/sendMessage`;
  const response = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      chat_id: chatId,
      text,
      disable_web_page_preview: true
    })
  });

  if (!response.ok) {
    const body = await response.text().catch(() => "");
    throw new Error(`telegram notify failed: ${response.status} ${body}`);
  }
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

  try {
    await sendTelegramMessage(lines.join("\n"));
  } catch (err) {
    console.warn("telegram notify failed", err);
  }
}

export async function notifyGoldenDogFound(payload: GoldenDogNotifyPayload): Promise<void> {
  const lines = [
    "EasyMeme Golden Dog detected",
    payload.tokenSymbol ? `Token: ${payload.tokenSymbol}` : "Token: (unknown)",
    `Address: ${payload.tokenAddress}`,
    typeof payload.goldenDogScore === "number" ? `GoldenDogScore: ${payload.goldenDogScore.toFixed(2)}` : "GoldenDogScore: (unknown)",
    typeof payload.riskScore === "number" ? `RiskScore: ${payload.riskScore.toFixed(2)}` : "RiskScore: (unknown)",
    payload.decisionReason ? `Decision: ${payload.decisionReason}` : "Decision: (none)",
  ];

  try {
    await sendTelegramMessage(lines.join("\n"));
  } catch (err) {
    console.warn("telegram notify failed", err);
  }
}
