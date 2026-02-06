export function RiskBadge({ level, score }: { level: string; score: number }) {
  const colors = {
    safe: 'bg-green-500',
    warning: 'bg-yellow-500',
    danger: 'bg-red-500',
  };

  const badge = {
    safe: 'SAFE',
    warning: 'WARN',
    danger: 'DANGER',
  };

  return (
    <div className="flex items-center gap-2">
      <span className="text-xs font-semibold">
        {badge[level as keyof typeof badge] || 'UNK'}
      </span>
      <span
        className={`px-2 py-0.5 rounded text-xs text-white ${
          colors[level as keyof typeof colors] || 'bg-gray-500'
        }`}
      >
        {score}
      </span>
    </div>
  );
}
