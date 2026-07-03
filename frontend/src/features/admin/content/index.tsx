import { Panel } from '../../shared/ui/admin';

// Honest stub: there is no CMS backend. Site copy ships as static files under src/content/.

export function AdminContent() {
  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">CMS</div>
          <h1>Content</h1>
        </div>
      </div>
      <Panel title="Not wired up yet">
        <p className="admin-empty">
          Journal posts, homepage blocks, and page management need a CMS backend. Today the
          site's editorial copy is maintained as static files in the frontend repo
          (src/content/). See the backend follow-ups in the tranche-3 spec.
        </p>
      </Panel>
    </>
  );
}
