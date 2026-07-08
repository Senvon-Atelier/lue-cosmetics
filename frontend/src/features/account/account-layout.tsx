import { Link, Outlet, useNavigate } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { Brand } from '../shared/ui/brand';
import { Icon } from '../shared/ui/icons';

// Ported from Lue/src/acct-pages.jsx (AcctSidebar) — real links only (spec §2.2):
// the mockup's tracking/subscriptions/returns/reviews/reorder/loyalty/referral/payments
// entries have no backend and are omitted.

export function AccountLayout() {
  const { user, isLoading, logout } = useAuth();
  const navigate = useNavigate();

  if (isLoading) {
    return (
      <div className="acct-layout">
        <div className="acct-loading">Loading…</div>
      </div>
    );
  }

  const displayName = user?.name || user?.email?.split('@')[0] || 'Member';
  const initial = displayName.charAt(0).toUpperCase();

  const handleSignOut = async () => {
    await logout();
    navigate({ to: '/' });
  };

  return (
    <div className="acct-layout">
      <aside className="acct-side">
        <div className="acct-side-brand">
          <Link to="/" className="brand" aria-label="Back to Lue home">
            <Brand />
          </Link>
        </div>
        <div className="acct-me">
          <div className="acct-avatar">{initial}</div>
          <div>
            <div className="acct-me-name">{displayName}</div>
            <div className="acct-me-tier">{user?.email}</div>
          </div>
        </div>
        <h5>Shop</h5>
        <Link
          to="/account"
          activeOptions={{ exact: true }}
          activeProps={{ className: 'active' }}
        >
          <span className="acct-side-link-label">
            <Icon name="user" size={14} /> Overview
          </span>
        </Link>
        <Link to="/account/orders" activeProps={{ className: 'active' }}>
          <span className="acct-side-link-label">
            <Icon name="bag" size={14} /> Orders
          </span>
        </Link>
        <Link to="/account/wishlist" activeProps={{ className: 'active' }}>
          <span className="acct-side-link-label">
            <Icon name="heart" size={14} /> Wishlist
          </span>
        </Link>
        <h5>Account</h5>
        <Link to="/account/addresses" activeProps={{ className: 'active' }}>
          <span className="acct-side-link-label">Addresses</span>
        </Link>
        <Link to="/account/settings" activeProps={{ className: 'active' }}>
          <span className="acct-side-link-label">Settings</span>
        </Link>
        <div className="acct-side-foot">
          <button onClick={handleSignOut}>
            <span className="acct-side-link-label">
              <Icon name="arrowLeft" size={14} /> Sign out
            </span>
          </button>
        </div>
      </aside>
      <Outlet />
    </div>
  );
}
