import Link from 'next/link';
import { getGoldenDogs } from '@/lib/api-server';

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
  const minScoreValue = Array.isArray(minScoreRaw) ? minScoreRaw[0] : minScoreRaw;
  const riskValue = Array.isArray(riskRaw) ? riskRaw[0] : riskRaw;
  const queryValue = Array.isArray(queryRaw) ? queryRaw[0] : queryRaw;
  const sortValue = Array.isArray(sortRaw) ? sortRaw[0] : sortRaw;
  const orderValue = Array.isArray(orderRaw) ? orderRaw[0] : orderRaw;
  const pageValue = Array.isArray(pageRaw) ? pageRaw[0] : pageRaw;
  const pageSizeValue = Array.isArray(pageSizeRaw) ? pageSizeRaw[0] : pageSizeRaw;
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
          <Link href="/" className="text-sm text-white/60 hover:text-white">
            ← 返回首页
          </Link>
          <h1 className="text-3xl font-semibold mt-2">金狗列表</h1>
          <p className="text-sm text-white/60 mt-1">
            按有效分数排序，已过滤 EXPIRED。
          </p>
        </div>
        <div className="flex flex-wrap gap-3">
          <div className="rounded-xl border border-white/20 px-4 py-2 text-xs text-white/70">
            共 {tokens.length} 个机会
          </div>
          <div className="rounded-xl border border-white/20 px-4 py-2 text-xs text-white/70">
            最高有效分数 {topScore}
          </div>
          <Link
            href="/ai-trades"
            className="rounded-xl border border-white/20 px-4 py-2 text-xs text-white/70 hover:text-white"
          >
            AI 交易历史
          </Link>
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
        />
        <TokenGrid tokens={pagedTokens} />
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
          }}
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
}: {
  minScore: number;
  risk: string;
  query: string;
  sort: string;
  order: string;
  pageSize: number;
}) {
  return (
    <form className="mb-6 flex flex-wrap items-center gap-3 text-sm">
      <label className="flex flex-col gap-1 text-white/70">
        关键词
        <input
          name="q"
          type="search"
          placeholder="Symbol / Name / Address"
          defaultValue={query || ''}
          className="w-52 rounded-xl border border-white/20 bg-transparent px-3 py-2 text-white focus:outline-none focus:border-white/50"
        />
      </label>
      <label className="flex flex-col gap-1 text-white/70">
        最小有效分数
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
        风险等级
        <select
          name="risk"
          className="w-36 rounded-xl border border-white/20 bg-transparent px-3 py-2 text-white focus:outline-none focus:border-white/50"
          defaultValue={risk}
        >
          <option value="all">全部</option>
          <option value="safe">SAFE</option>
          <option value="warning">WARNING</option>
          <option value="danger">DANGER</option>
        </select>
      </label>
      <label className="flex flex-col gap-1 text-white/70">
        排序
        <select
          name="sort"
          className="w-36 rounded-xl border border-white/20 bg-transparent px-3 py-2 text-white focus:outline-none focus:border-white/50"
          defaultValue={sort}
        >
          <option value="effective">有效分数</option>
          <option value="golden">金狗分数</option>
          <option value="risk">风险分数</option>
        </select>
      </label>
      <label className="flex flex-col gap-1 text-white/70">
        排序方向
        <select
          name="order"
          className="w-28 rounded-xl border border-white/20 bg-transparent px-3 py-2 text-white focus:outline-none focus:border-white/50"
          defaultValue={order}
        >
          <option value="desc">降序</option>
          <option value="asc">升序</option>
        </select>
      </label>
      <label className="flex flex-col gap-1 text-white/70">
        每页数量
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
      <button
        type="submit"
        className="mt-6 px-4 py-2 rounded-xl bg-white/10 text-white hover:bg-white/20"
      >
        筛选
      </button>
    </form>
  );
}

function TokenGrid({
  tokens,
}: {
  tokens: Awaited<ReturnType<typeof getGoldenDogs>>;
}) {
  return (
    <div className="grid gap-4">
      {tokens.map((token) => (
        <Link
          key={token.address}
          href={`/tokens/${token.address}`}
          className="rounded-2xl border border-white/10 bg-white/5 p-5 transition hover:border-white/30"
        >
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
              <p className="text-white/50 text-xs">Effective Score</p>
              <p className="text-white text-lg font-semibold">
                {token.effectiveScore}
              </p>
            </div>
            <div>
              <p className="text-white/50 text-xs">Golden Dog Score</p>
              <p className="text-white text-lg font-semibold">
                {token.goldenDogScore}
              </p>
            </div>
            <div>
              <p className="text-white/50 text-xs">Risk Score</p>
              <p className="text-white text-lg font-semibold">{token.riskScore}</p>
            </div>
            <div>
              <p className="text-white/50 text-xs">Time Decay</p>
              <p className="text-white text-lg font-semibold">
                {(token.timeDecayFactor * 100).toFixed(0)}%
              </p>
            </div>
          </div>
        </Link>
      ))}

      {tokens.length === 0 && (
        <div className="rounded-2xl border border-white/10 bg-white/5 p-10 text-center text-white/70">
          暂无金狗数据，请稍后再试。
        </div>
      )}
    </div>
  );
}

function Pagination({
  totalPages,
  page,
  searchParams,
}: {
  totalPages: number;
  page: number;
  searchParams: Record<string, string>;
}) {
  if (totalPages <= 1) return null;

  const buildQuery = (nextPage: number) => {
    const params = new URLSearchParams({ ...searchParams, page: String(nextPage) });
    return `/golden-dogs?${params.toString()}`;
  };

  return (
    <div className="mt-8 flex items-center justify-between text-sm text-white/70">
      <div>
        Page {page} / {totalPages}
      </div>
      <div className="flex items-center gap-3">
        <Link
          href={buildQuery(Math.max(1, page - 1))}
          className="px-3 py-2 rounded-xl border border-white/20 hover:border-white/50"
        >
          上一页
        </Link>
        <Link
          href={buildQuery(Math.min(totalPages, page + 1))}
          className="px-3 py-2 rounded-xl border border-white/20 hover:border-white/50"
        >
          下一页
        </Link>
      </div>
    </div>
  );
}
