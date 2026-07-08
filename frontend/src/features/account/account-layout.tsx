import { Link, Outlet, useNavigate } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { Icon } from '../shared/ui/icons';

export function AccountLayout() {
  const { user, isLoading, logout } = useAuth();
  const navigate = useNavigate();

  if (isLoading) {
    return (
      <div className="acct-page">
        <div className="wrap">
          <div className="acct-loading">Loading…</div>
        </div>
      </div>
    );
  }

  const displayName = user?.name || user?.email?.split('@')[0] || 'Member';

  const handleSignOut = async () => {
    await logout();
    navigate({ to: '/' });
  };

  return (
    <div className="acct-page">
      <div className="wrap">
        <div className="acct-top">
          <div>
            <div className="eyebrow" style={{ color: 'var(--lavender-700)' }}>Welcome back</div>
            <h1 className="h-display" style={{ fontSize: 'clamp(28px, 3.5vw, 48px)', margin: '4px 0 0' }}>
              Hello, {displayName}.
            </h1>
          </div>
          <button className="btn btn-ghost" onClick={handleSignOut} style={{ flexShrink: 0 }}>
            <Icon name="arrowLeft" size={14} /> Sign out
          </button>
        </div>

        <nav className="acct-nav">
          <Link to="/account" activeOptions={{ exact: true }} activeProps={{ className: 'active' }}>
            Overview
          </Link>
          <Link to="/account/orders" activeProps={{ className: 'active' }}>
            Orders
          </Link>
          <Link to="/account/wishlist" activeProps={{ className: 'active' }}>
            Wishlist
          </Link>
          <Link to="/account/addresses" activeProps={{ className: 'active' }}>
            Addresses
          </Link>
          <Link to="/account/settings" activeProps={{ className: 'active' }}>
            Settings
          </Link>
        </nav>

        <Outlet />
      </div>
    </div>
  );
}