import { useEffect } from 'react';
import { Outlet, Link, useNavigate } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { Button } from '../shared/ui/button';

export function AccountLayout() {
  const { user, isAuthenticated, isLoading } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      navigate({ to: '/login', replace: true });
    }
  }, [isAuthenticated, isLoading, navigate]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-ink-muted">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      {/* Header */}
      <div className="border-b border-line">
        <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
          <div className="flex items-center justify-between">
            <div>
              <h1 className="font-display text-2xl mb-1">My Account</h1>
              <p className="text-ink-muted text-sm">Welcome back, {user?.name || user?.email}</p>
            </div>
            <Button variant="outline" size="sm" onClick={() => navigate({ to: '/shop' })}>
              Continue Shopping
            </Button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
        <div className="grid md:grid-cols-4 gap-8">
          {/* Sidebar Navigation */}
          <nav className="md:col-span-1">
            <ul className="space-y-2">
              <li>
                <Link
                  to="/account"
                  className="block px-4 py-2 rounded-lg hover:bg-lavender-50 transition-colors font-label font-medium"
                  activeProps={{ className: 'bg-lavender-100 text-lavender-700' }}
                >
                  Dashboard
                </Link>
              </li>
              <li>
                <Link
                  to="/account/orders"
                  className="block px-4 py-2 rounded-lg hover:bg-lavender-50 transition-colors font-label font-medium"
                  activeProps={{ className: 'bg-lavender-100 text-lavender-700' }}
                >
                  Orders
                </Link>
              </li>
              <li>
                <Link
                  to="/account/addresses"
                  className="block px-4 py-2 rounded-lg hover:bg-lavender-50 transition-colors font-label font-medium"
                  activeProps={{ className: 'bg-lavender-100 text-lavender-700' }}
                >
                  Addresses
                </Link>
              </li>
              <li>
                <Link
                  to="/account/wishlist"
                  className="block px-4 py-2 rounded-lg hover:bg-lavender-50 transition-colors font-label font-medium"
                  activeProps={{ className: 'bg-lavender-100 text-lavender-700' }}
                >
                  Wishlist
                </Link>
              </li>
              <li>
                <Link
                  to="/account/settings"
                  className="block px-4 py-2 rounded-lg hover:bg-lavender-50 transition-colors font-label font-medium"
                  activeProps={{ className: 'bg-lavender-100 text-lavender-700' }}
                >
                  Settings
                </Link>
              </li>
            </ul>
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
