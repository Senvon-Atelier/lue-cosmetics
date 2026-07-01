interface KPICardProps {
  title: string;
  value: string | number;
  delta?: string;
  deltaDirection?: 'up' | 'down';
}

export function KPICard({ title, value, delta, deltaDirection }: KPICardProps) {
  return (
    <div className="bg-white border border-line rounded-xl p-5">
      <div className="text-[10px] uppercase tracking-wider text-ink-muted">{title}</div>
      <div className="font-display text-[32px] font-normal tracking-tight mt-2 mb-1">{value}</div>
      {delta && (
        <div
          className={`text-[11px] font-semibold ${
            deltaDirection === 'up' ? 'text-green-700' : deltaDirection === 'down' ? 'text-red-700' : ''
          }`}
        >
          {delta} vs. last period
        </div>
      )}
    </div>
  );
}
