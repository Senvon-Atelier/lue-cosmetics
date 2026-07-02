import { Icon } from '../shared/ui/icons';

const promises = [
  {
    icon: 'shield',
    title: '100% Authentic',
    description: 'Sourced direct from authorised distributors. If it\'s on our shelf, it\'s real.',
  },
  {
    icon: 'truck',
    title: 'Delivery across Ghana',
    description: 'Same-day in Accra over GHS 250. 2–4 days to other regions.',
  },
  {
    icon: 'heart',
    title: 'Concierge beauty',
    description: 'Routines built with you, in-store or on WhatsApp. No commission, just honesty.',
  },
  {
    icon: 'leaf',
    title: 'Clean, when we can',
    description: 'We prefer brands who publish their ingredient lists and mean them.',
  },
];

export function PromiseSection() {
  return (
    <section className="promise">
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '0 2rem' }}>
        <div className="promise-grid">
          {promises.map((promise) => (
            <div key={promise.title} className="promise-item">
              <div className="promise-icon">
                {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                <Icon name={promise.icon as any} size={22} />
              </div>
              <div className="promise-title">{promise.title}</div>
              <div className="promise-desc">{promise.description}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
