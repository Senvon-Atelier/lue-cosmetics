import { Panel } from '../../shared/ui/admin';

// Honest stub: campaigns, discount codes, and segments have no backend yet (spec §2.3, §8.4).

export function AdminMarketing() {
  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">Growth</div>
          <h1>Marketing</h1>
        </div>
      </div>
      <Panel title="Not wired up yet">
        <p className="admin-empty">
          Campaigns, discount codes, and customer segments need backend support before this
          page can show real data. See the backend follow-ups in the tranche-3 spec.
        </p>
      </Panel>
    </>
  );
}
