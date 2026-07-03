import type { ReactNode } from 'react';

// Ported from Rue/src/acct-pages.jsx (AcctHead, StatusPill)

export function AcctHead({
  eyebrow,
  title,
  children,
}: {
  eyebrow: string;
  title: string;
  children?: ReactNode;
}) {
  return (
    <div className="acct-head">
      <div>
        <div className="eyebrow">{eyebrow}</div>
        <h1>{title}</h1>
      </div>
      {children}
    </div>
  );
}

// Real order statuses (pending/paid/failed/cancelled) on the mockup's pill palette.
const KNOWN_STATUSES = new Set(['pending', 'paid', 'failed', 'cancelled']);

export function StatusPill({ status }: { status: string | undefined }) {
  if (!status) return null;
  const cls = KNOWN_STATUSES.has(status) ? ` status-${status}` : '';
  return (
    <span className={`status${cls}`}>
      <span className="status-dot"></span>
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  );
}
