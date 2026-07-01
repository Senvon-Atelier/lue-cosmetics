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
    <div className="section">
      <div className="wrap">
        <div className="grid grid-cols-1 lg:grid-cols-[240px_1fr] gap-12">
          {/* Sidebar Navigation */}
          <nav className="lg:sticky lg:top-24 lg:self-start h-fit">
            <div className="mb-6">
              <div className="eyebrow">Account</div>
              <h2 className="font-display text-xl">My Account</h2>
              <p className="text-ink-muted text-sm mt-1">Welcome back, {user?.name || user?.email}</p>
            </div>

            <ul className="space-y-1">
              <li>
                <Link
                  to="/account"
                  className="block px-4 py-3 rounded-full hover:bg-lavender-100 transition-colors duration-[var(--dur)] font-label font-medium text-sm"
                  activeProps={{ className: 'bg-lavender-600 text-paper hover:bg-lavender-700' }}
                >
                  Dashboard
                </Link>
              </li>
              <li>
                <Link
                  to="/account/orders"
                  className="block px-4 py-3 rounded-full hover:bg-lavender-100 transition-colors duration-[var(--dur)] font-label font-medium text-sm"
                  activeProps={{ className: 'bg-lavender-600 text-paper hover:bg-lavender-700' }}
                >
                  Orders
                </Link>
              </li>
              <li>
                <Link
                  to="/account/addresses"
                  className="block px-4 py-3 rounded-full hover:bg-lavender-100 transition-colors duration-[var(--dur)] font-label font-medium text-sm"
                  activeProps={{ className: 'bg-lavender-600 text-paper hover:bg-lavender-700' }}
                >
                  Addresses
                </Link>
              </li>
              <li>
                <Link
                  to="/account/wishlist"
                  className="block px-4 py-3 rounded-full hover:bg-lavender-100 transition-colors duration-[var(--dur)] font-label font-medium text-sm"
                  activeProps={{ className: 'bg-lavender-600 text-paper hover:bg-lavender-700' }}
                >
                  Wishlist
                </Link>
              </li>
              <li>
                <Link
                  to="/account/settings"
                  className="block px-4 py-3 rounded-full hover:bg-lavender-100 transition-colors duration-[var(--dur)] font-label font-medium text-sm"
                  activeProps={{ className: 'bg-lavender-600 text-paper hover:bg-lavender-700' }}
                >
                  Settings
                </Link>
              </li>
            </ul>
          </nav>

          {/* Content Area */}
          <main>
            <Outlet />
          </main>
        </div>
      </div>
    </div>
  );
}
