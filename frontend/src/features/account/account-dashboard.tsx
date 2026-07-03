import { useEffect, useState } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { getMeAddresses, getMeOrders } from '../../lib/api/generated/rueCosmeticsAPI';
import { useAuth } from '../../lib/auth/auth-provider';
import { formatGhs, formatOrderDate } from '../../lib/format/utils';
import { Icon } from '../shared/ui/icons';
import { AcctHead, StatusPill } from './acct-primitives';

type RecentOrder = {
  id?: string;
  status?: string;
  total_ghs?: number;
  created_at?: string;
};

type AddressSummary = { label?: string; is_default?: boolean };

export function AccountDashboard() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [recent, setRecent] = useState<RecentOrder[]>([]);
  const [ordersTotal, setOrdersTotal] = useState<number | null>(null);
  const [addresses, setAddresses] = useState<AddressSummary[] | null>(null);

  useEffect(() => {
    let cancelled = false;
    getMeOrders({ limit: 3 })
      .then((res) => {
        if (cancelled) return;
        setRecent(res.orders || []);
        setOrdersTotal(res.total ?? 0);
      })
      .catch(() => {
        if (!cancelled) setOrdersTotal(null);
      });
    getMeAddresses()
      .then((res) => {
        if (!cancelled) setAddresses(res.addresses || []);
      })
      .catch(() => {
        if (!cancelled) setAddresses(null);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const firstName = user?.name?.trim().split(/\s+/)[0];
  const defaultAddress = addresses?.find((a) => a.is_default);

  return (
    <main className="acct-main">
      <AcctHead
        eyebrow="Welcome back"
        title={firstName ? `Hello, ${firstName}.` : 'Hello.'}
      />

      <div className="acct-cards">
        <div className="acct-card">
          <div className="acct-card-k">Lifetime orders</div>
          <div className="acct-card-v">{ordersTotal ?? '—'}</div>
          <div className="acct-card-sub">All time</div>
        </div>
        <div className="acct-card">
          <div className="acct-card-k">Saved addresses</div>
          <div className="acct-card-v">{addresses ? addresses.length : '—'}</div>
          <div className="acct-card-sub">
            {defaultAddress?.label
              ? `${defaultAddress.label} set as default`
              : 'No default set'}
          </div>
        </div>
        <div className="acct-card acct-card-accent">
          <div className="acct-card-k">Member</div>
          <div className="acct-card-v">{firstName || 'Member'}</div>
          <div className="acct-card-sub">
            {user?.email} · {user?.email_verified ? 'Verified' : 'Unverified'}
          </div>
        </div>
      </div>

      <div className="acct-section">
        <div className="acct-section-head">
          <h2>Recent orders</h2>
          <Link className="auth-link" to="/account/orders">
            View all →
          </Link>
        </div>
        {recent.length === 0 ? (
          <div className="acct-empty">
            <p>No orders yet — when you place one, it will appear here.</p>
            <Link className="btn btn-primary" to="/shop">
              Start shopping
            </Link>
          </div>
        ) : (
          <div className="orders-table">
            <div className="orders-row head">
              <div>Order</div>
              <div>Date</div>
              <div>Total</div>
              <div>Status</div>
              <div></div>
            </div>
            {recent.map((o) => (
              <div key={o.id} className="orders-row">
                <div className="o-id">
                  #{(o.id || '').slice(0, 8).toUpperCase()}
                </div>
                <div>{formatOrderDate(o.created_at)}</div>
                <div className="price">{formatGhs(o.total_ghs || 0)}</div>
                <div>
                  <StatusPill status={o.status} />
                </div>
                <div>
                  <Link
                    className="link-btn"
                    to="/account/orders/$id"
                    params={{ id: o.id || '' }}
                  >
                    View
                  </Link>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="acct-section">
        <div className="acct-section-head">
          <h2>Quick links</h2>
        </div>
        <div className="dash-links">
          {[
            {
              to: '/account/orders',
              title: 'Orders',
              sub:
                ordersTotal !== null
                  ? `${ordersTotal} order${ordersTotal === 1 ? '' : 's'} placed`
                  : 'Your order history',
            },
            {
              to: '/account/addresses',
              title: 'Saved addresses',
              sub:
                addresses !== null
                  ? `${addresses.length} saved address${addresses.length === 1 ? '' : 'es'}`
                  : 'Manage delivery addresses',
            },
            {
              to: '/account/wishlist',
              title: 'Wishlist',
              sub: 'Saved items — coming soon',
            },
          ].map((l) => (
            <div
              key={l.to}
              className="dash-link-card"
              onClick={() => navigate({ to: l.to })}
            >
              <div>
                <div className="dash-link-card-title">{l.title}</div>
                <div className="dash-link-card-sub">{l.sub}</div>
              </div>
              <Icon name="arrow" size={16} />
            </div>
          ))}
        </div>
      </div>
    </main>
  );
}
