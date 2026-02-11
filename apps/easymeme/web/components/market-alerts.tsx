import { t, type Lang } from '@/lib/i18n';

type Props = {
  lang: Lang;
  alerts?: Array<Record<string, unknown>>;
};

function toNumber(value: unknown): number {
  if (typeof value === 'number') return value;
  if (typeof value === 'string') {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

export function MarketAlerts({ lang, alerts }: Props) {
  if (!alerts || alerts.length === 0) {
    return null;
  }

  return (
    <section className="rounded-2xl border border-white/10 bg-white/5 p-6">
      <h3 className="text-lg font-semibold text-white">{t(lang, 'token_alerts_title')}</h3>
      <div className="mt-4 grid gap-3">
        {alerts.map((alert, idx) => {
          const severity = String(alert.severity || 'MEDIUM').toUpperCase();
          const change = toNumber(alert.change) * 100;
          const tone = severity === 'HIGH' ? 'border-[#f07d7d]/50 bg-[#f07d7d]/10' : 'border-[#ffbf5c]/40 bg-[#ffbf5c]/10';
          return (
            <div key={`${alert.type || 'alert'}-${idx}`} className={`rounded-xl border px-4 py-3 ${tone}`}>
              <div className="text-sm font-semibold text-white">{String(alert.type || 'ALERT')}</div>
              <div className="mt-1 text-xs text-white/80">{String(alert.message || '')}</div>
              <div className="mt-1 text-xs text-white/70">{t(lang, 'token_alert_change')}: {change.toFixed(1)}%</div>
              <div className="mt-1 text-xs text-white/60">{String(alert.timestamp || '')}</div>
            </div>
          );
        })}
      </div>
    </section>
  );
}
