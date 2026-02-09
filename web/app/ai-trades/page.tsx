import Link from 'next/link';
import { getAITrades, getAITradeStats } from '@/lib/api-server';
import { resolveLang, t, withLang, type Lang } from '@/lib/i18n';
import { headers } from 'next/headers';

type AITradesPageProps = {
  searchParams?: { [key: string]: string | string[] | undefined };
};

export const dynamic = 'force-dynamic';

function formatPL(value: number) {
  const sign = value > 0 ? '+' : '';
  return `${sign}${value.toFixed(2)}%`;
}

function formatDate(value?: string | null) {
  if (!value) return 'N/A';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

function shortenHash(hash?: string | null) {
  if (!hash) return 'N/A';
  if (hash.length < 12) return hash;
  return `${hash.slice(0, 6)}...${hash.slice(-4)}`;
}

export default async function AITradesPage({ searchParams }: AITradesPageProps) {
  const lang = resolveLang(searchParams?.lang, headers().get('accept-language'));
  const statusRaw = searchParams?.status;
  const userRaw = searchParams?.user;
  const statusValue = Array.isArray(statusRaw) ? statusRaw[0] : statusRaw;
  const userValue = Array.isArray(userRaw) ? userRaw[0] : userRaw;
  const statusFilter =
    statusValue === 'success' || statusValue === 'failed' || statusValue === 'pending'
      ? statusValue
      : 'all';
  const userFilter = (userValue ?? '').toString().trim();

  const [trades, stats] = await Promise.all([getAITrades(50), getAITradeStats()]);
  const filteredTrades =
    statusFilter === 'all'
      ? trades
      : trades.filter((trade) => trade.status === statusFilter);
  const userFilteredTrades = userFilter
    ? filteredTrades.filter((trade) => trade.user_id === userFilter)
    : filteredTrades;

  return (
    <div className="min-h-screen px-6 pb-16">
      <header className="max-w-6xl mx-auto py-8 flex flex-wrap items-center justify-between gap-4">
        <div>
          <Link href={withLang('/', lang)} className="text-sm text-white/60 hover:text-white">
            {t(lang, 'trades_back_home')}
          </Link>
          <h1 className="text-3xl font-semibold mt-2">{t(lang, 'trades_title')}</h1>
          <p className="text-sm text-white/60 mt-1">{t(lang, 'trades_sub')}</p>
        </div>
        <Link
          href={withLang('/golden-dogs', lang)}
          className="rounded-xl border border-white/20 px-4 py-2 text-xs text-white/70 hover:text-white"
        >
          {t(lang, 'trades_view_golden')}
        </Link>
        <div className="flex items-center gap-2 text-xs text-white/60">
          <Link
            className={lang === 'zh' ? 'text-white' : 'hover:text-white'}
            href={withLang('/ai-trades', 'zh')}
          >
            中文
          </Link>
          <span>/</span>
          <Link
            className={lang === 'en' ? 'text-white' : 'hover:text-white'}
            href={withLang('/ai-trades', 'en')}
          >
            EN
          </Link>
        </div>
      </header>

      <main className="max-w-6xl mx-auto grid gap-6">
        <form className="flex flex-wrap items-center gap-3 text-sm">
          <label className="flex flex-col gap-1 text-white/70">
            {t(lang, 'trades_filters_user')}
            <input
              name="user"
              type="text"
              placeholder="default"
              defaultValue={userFilter}
              className="w-40 rounded-xl border border-white/20 bg-transparent px-3 py-2 text-white focus:outline-none focus:border-white/50"
            />
          </label>
          <label className="flex flex-col gap-1 text-white/70">
            {t(lang, 'trades_filters_status')}
            <select
              name="status"
              className="w-36 rounded-xl border border-white/20 bg-transparent px-3 py-2 text-white focus:outline-none focus:border-white/50"
              defaultValue={statusFilter}
            >
              <option value="all">{t(lang, 'trades_status_all')}</option>
              <option value="success">SUCCESS</option>
              <option value="pending">PENDING</option>
              <option value="failed">FAILED</option>
            </select>
          </label>
          <input type="hidden" name="lang" value={lang} />
          <button
            type="submit"
            className="mt-6 px-4 py-2 rounded-xl bg-white/10 text-white hover:bg-white/20"
          >
            {t(lang, 'trades_filters_submit')}
          </button>
        </form>

        <section className="grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border border-white/10 bg-white/5 p-6">
            <p className="text-xs text-white/50">{t(lang, 'trades_stat_total')}</p>
            <p className="text-2xl font-semibold text-white mt-2">{stats.count}</p>
          </div>
          <div className="rounded-2xl border border-white/10 bg-white/5 p-6">
            <p className="text-xs text-white/50">{t(lang, 'trades_stat_win')}</p>
            <p className="text-2xl font-semibold text-white mt-2">
              {(stats.winRate * 100).toFixed(1)}%
            </p>
          </div>
          <div className="rounded-2xl border border-white/10 bg-white/5 p-6">
            <p className="text-xs text-white/50">{t(lang, 'trades_stat_avg')}</p>
            <p className="text-2xl font-semibold text-white mt-2">
              {formatPL(stats.avgPL)}
            </p>
          </div>
        </section>

        <section className="grid gap-4 md:grid-cols-2">
          <div className="rounded-2xl border border-white/10 bg-white/5 p-6">
            <h2 className="text-lg font-semibold text-white mb-3">
              {t(lang, 'trades_stat_strategy')}
            </h2>
            <div className="grid gap-2 text-sm text-white/70">
              {stats.byStrategy.map((item) => (
                <div
                  key={item.strategy}
                  className="flex items-center justify-between border-b border-white/10 pb-2"
                >
                  <span>{item.strategy}</span>
                  <span className="text-white">
                    {(item.winRate * 100).toFixed(1)}% • {item.count} 笔 •{' '}
                    {formatPL(item.avgPL)}
                  </span>
                </div>
              ))}
              {stats.byStrategy.length === 0 && (
                <p className="text-white/50">N/A</p>
              )}
            </div>
          </div>
          <div className="rounded-2xl border border-white/10 bg-white/5 p-6">
            <h2 className="text-lg font-semibold text-white mb-3">
              {t(lang, 'trades_stat_period')}
            </h2>
            <div className="grid gap-2 text-sm text-white/70">
              {stats.byPeriod.map((item) => (
                <div
                  key={item.period}
                  className="flex items-center justify-between border-b border-white/10 pb-2"
                >
                  <span>{item.period}</span>
                  <span className="text-white">
                    {formatPL(item.totalPL)} • {(item.winRate * 100).toFixed(1)}%
                  </span>
                </div>
              ))}
              {stats.byPeriod.length === 0 && (
                <p className="text-white/50">N/A</p>
              )}
            </div>
          </div>
        </section>

        <section className="grid gap-4">
          {userFilteredTrades.map((trade) => (
            <div
              key={trade.id}
              className="rounded-2xl border border-white/10 bg-white/5 p-5"
            >
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h2 className="text-lg font-semibold">
                    {trade.token_symbol || 'UNKNOWN'}{' '}
                    <span className="text-white/50 text-sm">
                      {trade.token_address}
                    </span>
                  </h2>
                  <p className="text-xs text-white/60 mt-1">
                    {trade.type} • {formatDate(trade.timestamp)}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-xs text-white/50">{t(lang, 'trades_pl_label')}</p>
                  <p
                    className={`text-lg font-semibold ${
                      trade.profit_loss >= 0 ? 'text-[#7cf2a4]' : 'text-[#f07d7d]'
                    }`}
                  >
                    {formatPL(trade.profit_loss)}
                  </p>
                  <p className="text-xs text-white/50 mt-1">
                    {t(lang, 'trades_status_label')}: {trade.status || 'N/A'}
                  </p>
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-4 mt-4 text-sm text-white/70">
                <div>
                  <p className="text-white/50 text-xs">
                    {t(lang, 'trades_amount_in')} ({trade.type === 'BUY' ? 'BNB' : 'Token'})
                  </p>
                  <p className="text-white">{trade.amount_in || 'N/A'}</p>
                </div>
                <div>
                  <p className="text-white/50 text-xs">
                    {t(lang, 'trades_amount_out')} ({trade.type === 'BUY' ? 'Token' : 'BNB'})
                  </p>
                  <p className="text-white">{trade.amount_out || 'N/A'}</p>
                </div>
                <div>
                  <p className="text-white/50 text-xs">{t(lang, 'gd_golden')}</p>
                  <p className="text-white">{trade.golden_dog_score}</p>
                </div>
                <div>
                  <p className="text-white/50 text-xs">{t(lang, 'trades_strategy')}</p>
                  <p className="text-white">{trade.strategy_used || 'N/A'}</p>
                </div>
              </div>

              <div className="mt-4 grid gap-3 md:grid-cols-2 text-sm text-white/70">
                <div className="rounded-xl border border-white/10 bg-black/30 p-3">
                  <p className="text-xs uppercase tracking-widest text-white/50 mb-1">
                    {t(lang, 'trades_decision')}
                  </p>
                  <p>{trade.decision_reason || 'N/A'}</p>
                </div>
                <div className="rounded-xl border border-white/10 bg-black/30 p-3">
                  <p className="text-xs uppercase tracking-widest text-white/50 mb-1">
                    {t(lang, 'trades_info')}
                  </p>
                  <p>Gas: {trade.gas_used || 'N/A'}</p>
                  <p>Block: {trade.block_number || 'N/A'}</p>
                  {trade.error_message && (
                    <p className="text-xs text-[#f07d7d] mt-2">
                      Error: {trade.error_message}
                    </p>
                  )}
                  {trade.tx_hash ? (
                    <a
                      className="text-white/70 hover:text-white text-xs"
                      href={`https://bscscan.com/tx/${trade.tx_hash}`}
                      target="_blank"
                      rel="noreferrer"
                    >
                      BscScan: {shortenHash(trade.tx_hash)}
                    </a>
                  ) : (
                    <p className="text-xs text-white/50">BscScan: N/A</p>
                  )}
                </div>
              </div>
            </div>
          ))}

          {userFilteredTrades.length === 0 && (
            <div className="rounded-2xl border border-white/10 bg-white/5 p-10 text-center text-white/70">
              {t(lang, 'trades_empty')}
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
