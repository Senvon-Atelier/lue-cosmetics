interface StatusTagProps {
  status: string;
}

const KNOWN_TAGS = new Set([
  'paid', 'delivered', 'shipped', 'pending', 'processing', 'failed',
  'cancelled', 'fulfilled', 'live', 'draft', 'low', 'oos',
]);

export function StatusTag({ status }: StatusTagProps) {
  const lower = status.toLowerCase();
  const key = lower === 'low stock' ? 'low' : lower === 'out of stock' ? 'oos' : lower;
  const label =
    key === 'low' ? 'Low stock'
    : key === 'oos' ? 'Out of stock'
    : status.charAt(0).toUpperCase() + status.slice(1);
  const cls = KNOWN_TAGS.has(key) ? `tag-${key}` : 'tag-default';
  return <span className={`admin-tag ${cls}`}>{label}</span>;
}
