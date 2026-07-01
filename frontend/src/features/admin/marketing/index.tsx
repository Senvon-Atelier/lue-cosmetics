import { Panel } from '../../shared/ui/admin';

export function AdminMarketing() {
  return (
    <div>
      {/* Header */}
      <div className="flex items-start justify-between mb-7 gap-4 flex-wrap">
        <div>
          <div className="text-lavender-700 text-sm mb-1">Growth</div>
          <h1 className="font-display text-4xl font-normal">Marketing</h1>
        </div>
        <button className="px-3 py-2 bg-ink text-white rounded-lg text-sm font-semibold hover:bg-lavender-700 transition-colors">
          New campaign
        </button>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-4 gap-3 mb-6">
        {[
          ['Active campaigns', 8],
          ['Emails sent (7d)', '48.2k'],
          ['Open rate', '38.4%'],
          ['Active discount codes', 14],
        ].map(([label, value]) => (
          <div key={label} className="bg-white border border-line rounded-xl p-5">
            <div className="text-[10px] uppercase tracking-wider text-ink-muted">{label}</div>
            <div className="font-display text-[32px] font-normal tracking-tight mt-2">{value}</div>
          </div>
        ))}
      </div>

      {/* Active Campaigns */}
      <Panel title="Active campaigns">
        <table className="w-full border-collapse text-[13px]">
          <thead>
            <tr className="text-left">
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Campaign
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Type
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Audience
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Sent
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Open %
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Revenue
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Status
              </th>
            </tr>
          </thead>
          <tbody>
            {[
              ['Harmattan Edit launch', 'Email', 'All subscribers', '12,480', '42.1%', 'GH₵ 48,200', 'live'],
              ['Mid-year sale', 'Email + SMS', 'Bloom + Atelier', '3,218', '56.8%', 'GH₵ 82,400', 'live'],
              ['April giveaway', 'Landing page', 'Public', '—', '—', '—', 'live'],
              ['Abandoned cart', 'Automation', 'Triggered', '842', '28.4%', 'GH₵ 12,800', 'live'],
              ['Creator seeding Q2', 'Program', 'Influencers', '42', '—', '—', 'draft'],
            ].map(([name, type, audience, sent, open, revenue, status]) => (
              <tr key={name} className="hover:bg-[#FAFAFA] transition-colors">
                <td className="px-3 py-3 border-b border-line-soft font-display text-sm">{name}</td>
                <td className="px-3 py-3 border-b border-line-soft">{type}</td>
                <td className="px-3 py-3 border-b border-line-soft">{audience}</td>
                <td className="px-3 py-3 border-b border-line-soft font-variant-numeric tabular-nums">{sent}</td>
                <td className="px-3 py-3 border-b border-line-soft font-variant-numeric tabular-nums">{open}</td>
                <td className="px-3 py-3 border-b border-line-soft font-variant-numeric tabular-nums">{revenue}</td>
                <td className="px-3 py-3 border-b border-line-soft">
                  <span
                    className={`inline-flex px-2 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${
                      status === 'live' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'
                    }`}
                  >
                    {status}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Panel>

      {/* 2-column layout */}
      <div className="grid grid-cols-[2fr_1fr] gap-5 mt-5">
        {/* Discount Codes */}
        <Panel
          title="Discount codes"
          actions={<button className="px-3 py-2 border border-line rounded-lg text-sm font-semibold hover:bg-lavender-50">New code</button>}
        >
          <table className="w-full">
            <thead>
              <tr className="text-left border-b border-line">
                <th className="py-2 text-sm font-semibold">Code</th>
                <th className="py-2 text-sm font-semibold">Type</th>
                <th className="py-2 text-sm font-semibold">Used</th>
                <th className="py-2 text-sm font-semibold">Expires</th>
              </tr>
            </thead>
            <tbody>
              {[
                ['WELCOME10', '10% off', '2,142', 'Ongoing'],
                ['HARMATTAN', 'GH₵ 50 off', '418', 'May 3'],
                ['REFERRAL100', 'GH₵ 100', '287', 'Ongoing'],
                ['VIP25', '25% Atelier', '94', 'Ongoing'],
              ].map(([code, type, used, expires]) => (
                <tr key={code} className="border-b border-line-soft">
                  <td className="py-2 font-variant-numeric tabular-nums font-label font-bold">{code}</td>
                  <td className="py-2">{type}</td>
                  <td className="py-2 font-variant-numeric tabular-nums">{used}</td>
                  <td className="py-2 text-sm text-ink-muted">{expires}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Panel>

        {/* Segments */}
        <Panel title="Segments">
          <div className="grid gap-2">
            {[
              ['VIP — Atelier tier', '218', 'var(--ink)'],
              ['Lapsed — 90+ days', '842', 'var(--lavender-600)'],
              ['First-time buyer', '324', 'var(--lavender-400)'],
              ['High AOV (>GH₵ 500)', '612', 'var(--lavender-300)'],
            ].map(([name, count, color]) => (
              <div key={name} className="flex items-center justify-between gap-3 text-sm">
                <div className="flex items-center gap-2">
                  <span className="w-2.5 h-2.5 rounded-sm" style={{ background: color }} />
                  {name}
                </div>
                <strong>{count}</strong>
              </div>
            ))}
          </div>
        </Panel>
      </div>
    </div>
  );
}
