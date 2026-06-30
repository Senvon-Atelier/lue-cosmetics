import { useNavigate } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { Button } from '../shared/ui/button';

export function AccountDashboard() {
  const { user } = useAuth();
  const navigate = useNavigate();

  const quickStats = [
    {
      label: 'Orders',
      value: '0',
      action: () => navigate({ to: '/account/orders' }),
      actionText: 'View Orders',
    },
    {
      label: 'Wishlist',
      value: '0',
      action: () => navigate({ to: '/account/wishlist' }),
      actionText: 'View Wishlist',
    },
  ];

  return (
    <div>
      <div className="mb-8">
        <h2 className="font-display text-xl mb-2">Dashboard</h2>
        <p className="text-ink-muted">Manage your orders, addresses, and preferences.</p>
      </div>

      {/* Account Details */}
      <div className="bg-white rounded-lg p-6 mb-6" style={{ border: '1px solid var(--line)' }}>
        <h3 className="font-label font-semibold mb-4">Account Details</h3>
        <dl className="space-y-2 text-ink-soft">
          <div className="flex justify-between">
            <dt className="text-ink-muted">Name</dt>
            <dd className="font-medium">{user?.name || 'Not set'}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-ink-muted">Email</dt>
            <dd className="font-medium">{user?.email}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-ink-muted">Email Verified</dt>
            <dd className="font-medium">{user?.email_verified ? 'Yes' : 'No'}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-ink-muted">Account Type</dt>
            <dd className="font-medium capitalize">{user?.role || 'customer'}</dd>
          </div>
        </dl>
      </div>

      {/* Quick Stats */}
      <div className="grid grid-cols-2 gap-4 mb-6">
        {quickStats.map((stat) => (
          <div
            key={stat.label}
            className="bg-white rounded-lg p-6 text-center"
            style={{ border: '1px solid var(--line)' }}
          >
            <dt className="text-ink-muted text-sm mb-1">{stat.label}</dt>
            <dd className="text-3xl font-display text-ink mb-3">{stat.value}</dd>
            <Button
              variant="outline"
              size="sm"
              onClick={stat.action}
              className="w-full"
            >
              {stat.actionText}
            </Button>
          </div>
        ))}
      </div>

      {/* Quick Actions */}
      <div className="bg-lavender-50 rounded-lg p-6">
        <h3 className="font-label font-semibold mb-4">Quick Actions</h3>
        <div className="space-y-3">
          <Button
            variant="secondary"
            onClick={() => navigate({ to: '/shop' })}
            className="w-full"
          >
            Start Shopping
          </Button>
          <Button
            variant="outline"
            onClick={() => navigate({ to: '/account/addresses' })}
            className="w-full"
          >
            Manage Addresses
          </Button>
        </div>
      </div>
    </div>
  );
}
