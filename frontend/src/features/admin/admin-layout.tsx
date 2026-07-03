import { Outlet, Link } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';

const navSections = [
  {
    title: 'Overview',
    items: [
      { label: 'Dashboard', to: '/admin' },
      { label: 'Analytics', to: '/admin/analytics' },
    ],
  },
  {
    title: 'Commerce',
    items: [
      { label: 'Orders', to: '/admin/orders' },
      { label: 'Products', to: '/admin/products' },
      { label: 'Customers', to: '/admin/customers' },
    ],
  },
  {
    title: 'Growth',
    items: [
      { label: 'Marketing', to: '/admin/marketing' },
      { label: 'Content', to: '/admin/content' },
    ],
  },
  {
    title: 'System',
    items: [{ label: 'Settings', to: '/admin/settings' }],
  },
];

export function AdminLayout() {
  const { isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="admin-layout">
        <div className="admin-loading">Loading…</div>
      </div>
    );
  }

  return (
    <div className="admin-layout">
      <aside className="admin-side">
        <div className="admin-side-brand">
          <span className="brand-word">Rue</span>
          <span className="badge">Admin</span>
        </div>
        {navSections.map((section) => (
          <div key={section.title}>
            <h5>{section.title}</h5>
            {section.items.map((item) =>
              item.to === '/admin' ? (
                <Link
                  key={item.to}
                  to={item.to}
                  activeOptions={{ exact: true }}
                  activeProps={{ className: 'active' }}
                >
                  {item.label}
                </Link>
              ) : (
                <Link key={item.to} to={item.to} activeProps={{ className: 'active' }}>
                  {item.label}
                </Link>
              ),
            )}
          </div>
        ))}
        <div className="admin-side-foot">
          <Link to="/">← Storefront</Link>
        </div>
      </aside>
      <main className="admin-main">
        <Outlet />
      </main>
    </div>
  );
}
