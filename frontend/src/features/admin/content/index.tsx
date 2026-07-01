import { Panel } from '../../shared/ui/admin';

export function AdminContent() {
  return (
    <div>
      {/* Header */}
      <div className="flex items-start justify-between mb-7 gap-4 flex-wrap">
        <div>
          <div className="text-lavender-700 text-sm mb-1">CMS</div>
          <h1 className="font-display text-4xl font-normal">Content</h1>
        </div>
        <div className="flex gap-2">
          <button className="px-3 py-2 border border-line rounded-lg text-sm font-semibold hover:bg-lavender-50 transition-colors">
            Media library
          </button>
          <button className="px-3 py-2 bg-ink text-white rounded-lg text-sm font-semibold hover:bg-lavender-700 transition-colors">
            New post
          </button>
        </div>
      </div>

      {/* Journal Posts */}
      <Panel title="Journal posts">
        <table className="w-full border-collapse text-[13px]">
          <thead>
            <tr className="text-left">
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Title
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Author
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Category
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Views
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Updated
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Status
              </th>
            </tr>
          </thead>
          <tbody>
            {[
              ['How to layer actives without wrecking your barrier', 'Ama Owusu', 'Skincare', '8,420', '2d ago', 'live'],
              ['The Harmattan edit — behind the scenes', 'Editorial', 'Rituals', '4,210', '3d ago', 'live'],
              ['Shea from Tamale — a visit', 'Delali A.', 'Sourcing', '2,104', '5d ago', 'live'],
              ['Fragrance notes — a primer', 'Rue Atelier', 'Fragrance', '—', 'Today', 'draft'],
            ].map(([title, author, category, views, updated, status]) => (
              <tr key={title} className="hover:bg-[#FAFAFA] transition-colors">
                <td className="px-3 py-3 border-b border-line-soft font-display text-sm">{title}</td>
                <td className="px-3 py-3 border-b border-line-soft">{author}</td>
                <td className="px-3 py-3 border-b border-line-soft">{category}</td>
                <td className="px-3 py-3 border-b border-line-soft font-variant-numeric tabular-nums">{views}</td>
                <td className="px-3 py-3 border-b border-line-soft text-sm text-ink-muted">{updated}</td>
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
        {/* Homepage Blocks */}
        <Panel title="Homepage blocks">
          <div className="space-y-0">
            {['Hero — Harmattan Edit', 'Featured products rail', 'Journal preview', 'Testimonial carousel', 'Newsletter block'].map((block, i) => (
              <div
                key={block}
                className="flex justify-between items-center py-3 border-b border-line-soft last:border-0"
              >
                <div>
                  <div className="font-display text-sm">{block}</div>
                  <div className="text-xs text-ink-muted">Position {i + 1}</div>
                </div>
                <div className="flex gap-1">
                  <button className="px-3 py-1 border border-line rounded text-xs font-semibold hover:bg-lavender-50">Edit</button>
                  <button className="px-3 py-1 border border-line rounded text-xs hover:bg-lavender-50">⋮</button>
                </div>
              </div>
            ))}
          </div>
        </Panel>

        {/* Pages */}
        <Panel title="Pages">
          <div className="space-y-0">
            {['About', 'Shop', 'Journal', 'Contact', 'Loyalty', 'Affiliate'].map((page, i) => (
              <div
                key={page}
                className="flex justify-between items-center py-2 border-b border-line-soft last:border-0"
              >
                <div>{page}</div>
                <span className="inline-flex px-2 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider bg-green-100 text-green-700">
                  live
                </span>
              </div>
            ))}
          </div>
        </Panel>
      </div>
    </div>
  );
}
