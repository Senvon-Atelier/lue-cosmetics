import { Panel } from '../../shared/ui/admin';

// Read-only true facts; editable settings need backend support (spec §2.3, §8.4).

export function AdminSettings() {
  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">Configuration</div>
          <h1>Settings</h1>
        </div>
      </div>
      <Panel title="Store">
        <div className="kv-row">
          <span>Store</span>
          <span>Lue Cosmetics</span>
        </div>
        <div className="kv-row">
          <span>Currency</span>
          <span>GHS — Ghanaian cedi</span>
        </div>
        <div className="kv-row">
          <span>Payments</span>
          <span>Paystack</span>
        </div>
      </Panel>
      <Panel title="Editable settings">
        <p className="admin-empty">
          Store configuration lives in the backend deployment (env + config files); editable
          settings need backend support. See the backend follow-ups in the tranche-3 spec.
        </p>
      </Panel>
    </>
  );
}
