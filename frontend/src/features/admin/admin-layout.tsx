import { useEffect } from 'react';
import { Outlet, Link, useNavigate } from '@tanstack/react-router';
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
    items: [
      { label: 'Settings', to: '/admin/settings' },
    ],
  },
];

export function AdminLayout() {
  const { user, isAuthenticated, isAdmin, isLoading } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
  if (!isLoading && !isAuthenticated) {
    navigate({ to: '/login', replace: true });
    return;
  }

  if (!isLoading && isAuthenticated && !isAdmin) {
    navigate({ to: '/', replace: true });
  }
}, [isAuthenticated, isAdmin, isLoading, navigate]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-ink-muted">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
  return (
    <div className="min-h-screen flex items-center justify-center">
      Redirecting to login...
    </div>
  );
}

  if (!isAdmin) {
    return (
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
        <h1 className="font-display text-4xl mb-4">Access Denied</h1>
        <p className="text-ink-muted">You don't have permission to access the admin dashboard.</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      {/* Header */}
      <div className="border-b border-line">
        <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
          <div className="flex items-center justify-between">
            <div>
              <h1 className="font-display text-2xl mb-1">Admin Dashboard</h1>
              <p className="text-ink-muted text-sm">Welcome, {user?.name || user?.email}</p>
            </div>
            <button
              onClick={() => navigate({ to: '/shop' })}
              className="px-4 py-2 border border-line rounded-lg hover:bg-lavender-50 transition-colors font-label font-medium text-sm"
            >
              Back to Store
            </button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
        <div className="grid md:grid-cols-4 gap-8">
          {/* Sidebar Navigation */}
          <nav className="md:col-span-1">
            {navSections.map((section) => (
              <div key={section.title} className="mb-6">
                <h3 className="font-label font-medium text-ink-muted uppercase text-xs mb-2 tracking-wide">
                  {section.title}
                </h3>
                <ul className="space-y-1">
                  {section.items.map((item) => (
                    <li key={item.to}>
                      <Link
                        to={item.to}
                        className="block px-4 py-2 rounded-lg hover:bg-lavender-50 transition-colors font-label font-medium text-sm"
                        activeProps={{ className: 'bg-lavender-100 text-lavender-700' }}
                      >
                        {item.label}
                      </Link>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </nav>

          {/* Content Area */}
          <main className="md:col-span-3">
            <Outlet />
          </main>
        </div>
      </div>
    </div>
  );
}
