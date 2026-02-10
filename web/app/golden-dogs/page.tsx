import Link from 'next/link';
import { getGoldenDogs } from '@/lib/api-server';
import { resolveLang, t, withLang, type Lang } from '@/lib/i18n';
import { headers } from 'next/headers';

type GoldenDogsPageProps = {
  searchParams?: { [key: string]: string | string[] | undefined };
};

export const dynamic = 'force-dynamic';

function phaseTone(phase: string) {
  switch (phase) {
    case 'EARLY':
      return 'bg-[#7cf2a4] text-black';
    case 'PEAK':
      return 'bg-[#ffbf5c] text-black';
    case 'DECLINING':
      return 'bg-[#f07d7d] text-black';
    default:
      return 'bg-white/10 text-white';
  }
}

function riskTone(level: string) {
  switch (level) {
    case 'safe':
      return 'bg-[#7cf2a4] text-black';
    case 'warning':
      return 'bg-[#ffbf5c] text-black';
    case 'danger':
      return 'bg-[#f07d7d] text-black';
    default:
      return 'bg-white/10 text-white';
  }
}

export default async function GoldenDogsPage({ searchParams }: GoldenDogsPageProps) {
  const lang = resolveLang(searchParams?.lang, headers().get('accept-language'));
  const tokens = await getGoldenDogs(60);
  const topScore = tokens.reduce(
    (max, token) => Math.max(max, token.effectiveScore),
    0,
  );
  const minScoreRaw = searchParams?.minScore;
  const riskRaw = searchParams?.risk;
  const queryRaw = searchParams?.q;
  const sortRaw = searchParams?.sort;
  const orderRaw = searchParams?.order;
  const pageRaw = searchParams?.page;
  const pageSizeRaw = searchParams?.pageSize;
  const langRaw = searchParams?.lang;
  const minScoreValue = Array.isArray(minScoreRaw) ? minScoreRaw[0] : minScoreRaw;
  const riskValue = Array.isArray(riskRaw) ? riskRaw[0] : riskRaw;
  const queryValue = Array.isArray(queryRaw) ? queryRaw[0] : queryRaw;
  const sortValue = Array.isArray(sortRaw) ? sortRaw[0] : sortRaw;
  const orderValue = Array.isArray(orderRaw) ? orderRaw[0] : orderRaw;
  const pageValue = Array.isArray(pageRaw) ? pageRaw[0] : pageRaw;
  const pageSizeValue = Array.isArray(pageSizeRaw) ? pageSizeRaw[0] : pageSizeRaw;
  const langValue = Array.isArray(langRaw) ? langRaw[0] : langRaw;
  const minScore = Math.max(0, Math.min(100, Number(minScoreValue) || 0));
  const riskFilter =
    riskValue === 'safe' || riskValue === 'warning' || riskValue === 'danger'
      ? riskValue
      : 'all';
  const searchQuery = (queryValue ?? '').toString().trim().toLowerCase();
  const sortKey =
    sortValue === 'golden' || sortValue === 'risk' || sortValue === 'effective'
      ? sortValue
      : 'effective';
  const sortOrder = orderValue === 'asc' ? 'asc' : 'desc';
  const pageSize = Math.max(5, Math.min(30, Number(pageSizeValue) || 10));
  const page = Math.max(1, Number(pageValue) || 1);
  const filteredTokens = tokens.filter((token) => {
    if (token.effectiveScore < minScore) return false;
    if (riskFilter !== 'all' && token.riskLevel !== riskFilter) return false;
    if (searchQuery) {
      const blob = `${token.symbol ?? ''} ${token.name ?? ''} ${token.address}`
        .toLowerCase()
        .trim();
      if (!blob.includes(searchQuery)) return false;
    }
    return true;
  });
  const sortedTokens = [...filteredTokens].sort((a, b) => {
    const aVal =
      sortKey === 'golden'
        ? a.goldenDogScore
        : sortKey === 'risk'
          ? a.riskScore
          : a.effectiveScore;
    const bVal =
      sortKey === 'golden'
        ? b.goldenDogScore
        : sortKey === 'risk'
          ? b.riskScore
          : b.effectiveScore;
    if (aVal === bVal) {
      return a.analyzedAt && b.analyzedAt
        ? a.analyzedAt.localeCompare(b.analyzedAt)
        : 0;
    }
    return sortOrder === 'asc' ? aVal - bVal : bVal - aVal;
  });
  const totalPages = Math.max(1, Math.ceil(sortedTokens.length / pageSize));
  const safePage = Math.min(page, totalPages);
  const start = (safePage - 1) * pageSize;
  const pagedTokens = sortedTokens.slice(start, start + pageSize);

  return (
    <div className="min-h-screen px-6 pb-16">
      <header className="max-w-6xl mx-auto py-8 flex flex-wrap items-center justify-between gap-4">
        <div>
          <Link href={withLang('/', lang)} className="text-sm text-white/60 hover:text-white">
            {t(lang, 'gd_back_home')}
          </Link>
          <h1 className="text-3xl font-semibold mt-2">{t(lang, 'gd_title')}</h1>
          <p className="text-sm text-white/60 mt-1">
            {t(lang, 'gd_sub')}
          </p>
        </div>
        <div className="flex flex-wrap gap-3">
          <div className="rounded-xl border border-white/20 px-4 py-2 text-xs text-white/70">
            {t(lang, 'gd_total', { count: tokens.length })}
          </div>
          <div className="rounded-xl border border-white/20 px-4 py-2 text-xs text-white/70">
            {t(lang, 'gd_top', { score: topScore })}
          </div>
          <Link
            href={withLang('/ai-trades', lang)}
            className="rounded-xl border border-white/20 px-4 py-2 text-xs text-white/70 hover:text-white"
          >
            {t(lang, 'nav_trades')}
          </Link>
          <div className="flex items-center gap-2 text-xs text-white/60">
            <Link
              className={lang === 'zh' ? 'text-white' : 'hover:text-white'}
              href={withLang('/golden-dogs', 'zh')}
            >
              中文
            </Link>
            <span>/</span>
            <Link
              className={lang === 'en' ? 'text-white' : 'hover:text-white'}
              href={withLang('/golden-dogs', 'en')}
            >
              EN
            </Link>
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto">
        <Filters
          minScore={minScore}
          risk={riskFilter}
          query={searchQuery}
          sort={sortKey}
          order={sortOrder}
          pageSize={pageSize}
          lang={lang}
        />
        <TokenGrid tokens={pagedTokens} lang={lang} />
        <Pagination
          totalPages={totalPages}
          page={safePage}
          searchParams={{
            minScore: String(minScore || ''),
            risk: riskFilter,
            q: searchQuery,
            sort: sortKey,
            order: sortOrder,
            pageSize: String(pageSize),
            lang: langValue || lang,
          }}
          lang={lang}
        />
      </main>
    </div>
  );
}

function Filters({
  minScore,
  risk,
  query,
  sort,
  order,
  pageSize,
  lang,
}: {
  minScore: number;
  risk: string;
  query: string;
  sort: string;
  order: string;
  pageSize: number;
  lang: Lang;
}) {
  return (
    <form className="mb-6 flex flex-wrap items-center gap-3 text-sm">
      <label className="flex flex-col gap-1 text-white/70">
        {t(lang, 'gd_filters_keyword')}
        <input
          name="q"
          type="search"
          placeholder={t(lang, 'gd_filters_placeholder')}
          defaultValue={query || ''}
          className="w-52 rounded-xl border border-white/20 bg-transparent px-3 py-2 text-white focus:outline-none focus:border-white/50"
        />
      </label>
      <label className="flex flex-col gap-1 text-white/70">
        {t(lang, 'gd_filters_min_score')}
        <input
          name="minScore"
          type="number"
          min={0}
          max={100}
          placeholder="0"
          defaultValue={minScore || ''}
          className="w-32 rounded-xl border border-white/20 bg-transparent px-3 py-2 text-white focus:outline-none focus:border-white/50"
        />
      </label>
      <label className="flex flex-col gap-1 text-white/70">
        {t(lang, 'gd_filters_risk')}
        <select
          name="risk"
          className="w-36 rounded-xl border border-white/20 bg-transparent px-3 py-2 text-white focus:outline-none focus:border-white/50"
          defaultValue={risk}
        >
          <option value="all">{t(lang, 'gd_filters_all')}</option>
          <option value="safe">SAFE</option>
          <option value="warning">WARNING</option>
          <option value="danger">DANGER</option>
        </select>
      </label>
      <label className="flex flex-col gap-1 text-white/70">
        {t(lang, 'gd_filters_sort')}
        <select
          name="sort"
          className="w-36 rounded-xl border border-white/20 bg-transparent px-3 py-2 text-white focus:outline-none focus:border-white/50"
          defaultValue={sort}
        >
          <option value="effective">{t(lang, 'gd_sort_effective')}</option>
          <option value="golden">{t(lang, 'gd_sort_golden')}</option>
          <option value="risk">{t(lang, 'gd_sort_risk')}</option>
        </select>
      </label>
      <label className="flex flex-col gap-1 text-white/70">
        {t(lang, 'gd_filters_order')}
        <select
          name="order"
          className="w-28 rounded-xl border border-white/20 bg-transparent px-3 py-2 text-white focus:outline-none focus:border-white/50"
          defaultValue={order}
        >
          <option value="desc">{t(lang, 'gd_order_desc')}</option>
          <option value="asc">{t(lang, 'gd_order_asc')}</option>
        </select>
      </label>
      <label className="flex flex-col gap-1 text-white/70">
        {t(lang, 'gd_filters_page_size')}
        <select
          name="pageSize"
          className="w-24 rounded-xl border border-white/20 bg-transparent px-3 py-2 text-white focus:outline-none focus:border-white/50"
          defaultValue={pageSize}
        >
          <option value="10">10</option>
          <option value="15">15</option>
          <option value="20">20</option>
          <option value="30">30</option>
        </select>
      </label>
      <input type="hidden" name="lang" value={lang} />
      <button
        type="submit"
        className="mt-6 px-4 py-2 rounded-xl bg-white/10 text-white hover:bg-white/20"
      >
        {t(lang, 'gd_filters_submit')}
      </button>
    </form>
  );
}

function TokenGrid({
  tokens,
  lang,
}: {
  tokens: Awaited<ReturnType<typeof getGoldenDogs>>;
  lang: Lang;
}) {
  return (
    <div className="grid gap-4">
      {tokens.map((token) => (
        <div
          key={token.address}
          className="rounded-2xl border border-white/10 bg-white/5 p-5 transition hover:border-white/30"
        >
          <Link href={withLang(`/tokens/${token.address}`, lang)} className="block">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 className="text-lg font-semibold">
                  {token.symbol || 'UNKNOWN'}{' '}
                  <span className="text-white/50 text-sm">{token.name}</span>
                </h2>
                <p className="text-xs text-white/60 mt-1">{token.address}</p>
              </div>
              <div className="flex items-center gap-2 text-xs font-semibold">
                <span className={`px-3 py-1 rounded-full ${riskTone(token.riskLevel)}`}>
                  {token.riskLevel.toUpperCase()}
                </span>
                <span className={`px-3 py-1 rounded-full ${phaseTone(token.phase)}`}>
                  {token.phase}
                </span>
              </div>
            </div>
            <div className="grid gap-4 md:grid-cols-4 mt-4 text-sm text-white/70">
              <div>
                <p className="text-white/50 text-xs">{t(lang, 'gd_effective')}</p>
                <p className="text-white text-lg font-semibold">
                  {token.effectiveScore}
                </p>
              </div>
              <div>
                <p className="text-white/50 text-xs">{t(lang, 'gd_golden')}</p>
                <p className="text-white text-lg font-semibold">
                  {token.goldenDogScore}
                </p>
              </div>
              <div>
                <p className="text-white/50 text-xs">{t(lang, 'gd_risk')}</p>
                <p className="text-white text-lg font-semibold">{token.riskScore}</p>
              </div>
              <div>
                <p className="text-white/50 text-xs">{t(lang, 'gd_decay')}</p>
                <p className="text-white text-lg font-semibold">
                  {(token.timeDecayFactor * 100).toFixed(0)}%
                </p>
              </div>
            </div>
          </Link>
          <div className="mt-4">
            <a
              href={`https://gmgn.ai/bsc/token/${token.address}`}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center rounded-lg border border-white/20 px-3 py-1.5 text-xs text-white/80 hover:text-white hover:border-white/40"
            >
              GMGN
            </a>
          </div>
        </div>
      ))}

      {tokens.length === 0 && (
        <div className="rounded-2xl border border-white/10 bg-white/5 p-10 text-center text-white/70">
          {t(lang, 'gd_empty')}
        </div>
      )}
    </div>
  );
}

function Pagination({
  totalPages,
  page,
  searchParams,
  lang,
}: {
  totalPages: number;
  page: number;
  searchParams: Record<string, string>;
  lang: Lang;
}) {
  if (totalPages <= 1) return null;

  const buildQuery = (nextPage: number) => {
    const params = new URLSearchParams({ ...searchParams, page: String(nextPage) });
    return `/golden-dogs?${params.toString()}`;
  };

  return (
    <div className="mt-8 flex items-center justify-between text-sm text-white/70">
      <div>
        {t(lang, 'page', { page, total: totalPages })}
      </div>
      <div className="flex items-center gap-3">
        <Link
          href={buildQuery(Math.max(1, page - 1))}
          className="px-3 py-2 rounded-xl border border-white/20 hover:border-white/50"
        >
          {t(lang, 'prev')}
        </Link>
        <Link
          href={buildQuery(Math.min(totalPages, page + 1))}
          className="px-3 py-2 rounded-xl border border-white/20 hover:border-white/50"
        >
          {t(lang, 'next')}
        </Link>
      </div>
    </div>
  );
}
