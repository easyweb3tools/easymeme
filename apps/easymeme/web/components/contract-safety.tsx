import { CopyButton } from '@/components/copy-button';
import { t, type Lang } from '@/lib/i18n';

type Props = {
  lang: Lang;
  goplus?: Record<string, unknown>;
};

function toBool(value: unknown): boolean {
  if (typeof value === 'boolean') return value;
  if (typeof value === 'number') return value !== 0;
  if (typeof value === 'string') {
    const v = value.trim().toLowerCase();
    return v === '1' || v === 'true' || v === 'yes';
  }
  return false;
}

function toNum(value: unknown): number {
  if (typeof value === 'number') return value;
  if (typeof value === 'string') {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

function renderCheck(pass: boolean, label: string, trueLabel: string, falseLabel: string) {
  return (
    <div className="rounded-xl border border-white/10 bg-black/20 px-3 py-2 text-sm">
      <span className={pass ? 'text-[#7cf2a4]' : 'text-[#f07d7d]'}>{pass ? '✅' : '❌'}</span>{' '}
      <span className="text-white/80">{label}:</span>{' '}
      <span className="text-white">{pass ? trueLabel : falseLabel}</span>
    </div>
  );
}

export function ContractSafety({ lang, goplus }: Props) {
  if (!goplus || Object.keys(goplus).length === 0) {
    return <section className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-white/60">{t(lang, 'token_contract_safety_empty')}</section>;
  }

  const creator = (goplus.creator_address as string) || 'N/A';
  const checks = [
    { pass: !toBool(goplus.is_honeypot), label: t(lang, 'token_safety_honeypot'), yes: t(lang, 'token_no'), no: t(lang, 'token_yes') },
    { pass: toBool(goplus.is_open_source), label: t(lang, 'token_safety_open_source'), yes: t(lang, 'token_yes'), no: t(lang, 'token_no') },
    { pass: !toBool(goplus.is_mintable), label: t(lang, 'token_safety_mintable'), yes: t(lang, 'token_no'), no: t(lang, 'token_yes') },
    { pass: !toBool(goplus.is_proxy), label: t(lang, 'token_safety_proxy'), yes: t(lang, 'token_no'), no: t(lang, 'token_yes') },
    { pass: !toBool(goplus.can_take_back_ownership), label: t(lang, 'token_safety_take_back_ownership'), yes: t(lang, 'token_no'), no: t(lang, 'token_yes') },
  ];

  return (
    <section className="rounded-2xl border border-white/10 bg-white/5 p-6">
      <h3 className="text-lg font-semibold text-white">{t(lang, 'token_contract_safety_title')}</h3>
      <div className="mt-4 grid gap-2 md:grid-cols-2">
        {checks.map((c) => <div key={c.label}>{renderCheck(c.pass, c.label, c.yes, c.no)}</div>)}
      </div>
      <div className="mt-4 grid gap-2 text-sm text-white/80 md:grid-cols-2">
        <div>{t(lang, 'token_safety_holders')}: {toNum(goplus.holder_count)}</div>
        <div>{t(lang, 'token_safety_lp_holders')}: {toNum(goplus.lp_holder_count)}</div>
      </div>
      <div className="mt-3 flex items-center gap-2 text-sm text-white/80">
        <span>{t(lang, 'token_creator')}: {creator}</span>
        {creator !== 'N/A' && <CopyButton value={creator} label={t(lang, 'copy')} copiedLabel={t(lang, 'copied')} />}
      </div>
    </section>
  );
}
