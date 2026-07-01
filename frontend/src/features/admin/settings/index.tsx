import { Panel } from '../../shared/ui/admin';

export function AdminSettings() {
  return (
    <div>
      {/* Header */}
      <div className="mb-7">
        <div className="text-lavender-700 text-sm mb-1">Configuration</div>
        <h1 className="font-display text-4xl font-normal">Settings</h1>
      </div>

      {/* 2-column layout */}
      <div className="grid grid-cols-[2fr_1fr] gap-5">
        {/* Left Column */}
        <div className="space-y-5">
          {/* Store Details */}
          <Panel title="Store details">
            <div className="grid gap-3">
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1">
                  <label className="text-[10px] font-bold uppercase tracking-wider text-ink-muted">Store name</label>
                  <input type="text" defaultValue="Rue Cosmetics" className="px-3 py-2 border border-line rounded-lg text-sm bg-white focus:outline-none focus:border-lavender-400" />
                </div>
                <div className="flex flex-col gap-1">
                  <label className="text-[10px] font-bold uppercase tracking-wider text-ink-muted">Trading name</label>
                  <input type="text" defaultValue="Rue Cosmetics Ltd." className="px-3 py-2 border border-line rounded-lg text-sm bg-white focus:outline-none focus:border-lavender-400" />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1">
                  <label className="text-[10px] font-bold uppercase tracking-wider text-ink-muted">Currency</label>
                  <select className="px-3 py-2 border border-line rounded-lg text-sm bg-white focus:outline-none focus:border-lavender-400">
                    <option>GHS — Ghanaian Cedi</option>
                  </select>
                </div>
                <div className="flex flex-col gap-1">
                  <label className="text-[10px] font-bold uppercase tracking-wider text-ink-muted">Timezone</label>
                  <select className="px-3 py-2 border border-line rounded-lg text-sm bg-white focus:outline-none focus:border-lavender-400">
                    <option>Africa/Accra (GMT+0)</option>
                  </select>
                </div>
              </div>
              <div className="flex flex-col gap-1">
                <label className="text-[10px] font-bold uppercase tracking-wider text-ink-muted">Registered address</label>
                <input type="text" defaultValue="14 Oxford St, Osu, Accra, Ghana" className="px-3 py-2 border border-line rounded-lg text-sm bg-white focus:outline-none focus:border-lavender-400" />
              </div>
            </div>
          </Panel>

          {/* Payments */}
          <Panel title="Payments">
            <div className="space-y-0">
              {[
                ['Paystack', 'Live', 'Cards + MoMo'],
                ['MTN MoMo', 'Live', 'Direct'],
                ['Vodafone Cash', 'Disabled', 'Direct'],
              ].map(([name, status, description]) => (
                <div
                  key={name}
                  className="flex justify-between items-center py-3 border-b border-line-soft last:border-0"
                >
                  <div>
                    <div className="font-display text-sm">{name}</div>
                    <div className="text-xs text-ink-muted">{description}</div>
                  </div>
                  <span className={`inline-flex px-2 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${
                    status === 'Live' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'
                  }`}>
                    {status}
                  </span>
                </div>
              ))}
            </div>
          </Panel>
        </div>

        {/* Right Column */}
        <div className="space-y-5">
          {/* Team */}
          <Panel
            title="Team"
            actions={<button className="px-3 py-2 border border-line rounded-lg text-sm font-semibold hover:bg-lavender-50">Invite</button>}
          >
            <div className="space-y-0">
              {[
                ['Ama Owusu', 'Owner'],
                ['Delali A.', 'Admin'],
                ['Kofi M.', 'Fulfilment'],
                ['Esi O.', 'Editor'],
              ].map(([name, role]) => (
                <div
                  key={name}
                  className="flex justify-between items-center py-2 border-b border-line-soft last:border-0"
                >
                  <div className="flex items-center gap-2">
                    <div className="w-8 h-8 rounded-full bg-lavender-300 flex items-center justify-center font-display italic">
                      {name[0]}
                    </div>
                    <div>{name}</div>
                  </div>
                  <span className="inline-flex px-2 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider bg-green-100 text-green-700">
                    {role}
                  </span>
                </div>
              ))}
            </div>
          </Panel>

          {/* Integrations */}
          <Panel title="Integrations">
            <div className="space-y-0">
              {['Klaviyo · Email', 'Meta Pixel · Ads', 'Google Analytics 4', 'DHL · Shipping', 'Trustpilot · Reviews'].map((integration, i) => (
                <div
                  key={integration}
                  className="flex justify-between items-center py-2 border-b border-line-soft last:border-0"
                >
                  <div className="text-sm">{integration}</div>
                  <span className="inline-flex px-2 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider bg-green-100 text-green-700">
                    live
                  </span>
                </div>
              ))}
            </div>
          </Panel>
        </div>
      </div>
    </div>
  );
}
