import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useGetAdminCustomers } from '../../../lib/api/generated/rueCosmeticsAPI';
import { Panel } from '../../shared/ui/admin';

export function AdminCustomers() {
  const navigate = useNavigate();
  const [page, setPage] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');
  const [tierFilter, setTierFilter] = useState('');

  const { data: customersData, isLoading, error } = useGetAdminCustomers({
    page: page + 1,
    page_size: 20,
  });

  const formatCurrency = (amount: number) => {
    return `GH₵${(amount / 100).toLocaleString()}`;
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return '—';

    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const totalPages = customersData?.total_pages ?? 1;
  const customers = customersData?.customers ?? [];

  // Filter customers based on search (client-side for now)
  const filteredCustomers = customers.filter((c) => {
    if (searchQuery && !c.name?.toLowerCase().includes(searchQuery.toLowerCase()) && !(c.email ?? '')
      .toLowerCase()
      .includes(searchQuery.toLowerCase())) {
      return false;
    }
    return true;
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-ink-muted">Loading customers...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg">
        Failed to load customers: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  return (
    <div>
      {/* Header */}
      <div className="flex items-start justify-between mb-7 gap-4 flex-wrap">
        <div>
          <div className="text-lavender-700 text-sm mb-1">CRM</div>
          <h1 className="font-display text-4xl font-normal">Customers</h1>
        </div>
        <div className="flex gap-2">
          <button className="px-3 py-2 border border-line rounded-lg text-sm font-semibold hover:bg-lavender-50 transition-colors">
            Create segment
          </button>
          <button className="px-3 py-2 bg-ink text-white rounded-lg text-sm font-semibold hover:bg-lavender-700 transition-colors">
            Export list
          </button>
        </div>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-4 gap-3 mb-6">
        {[
          ['Total', customers.length],
          ['New (30d)', customers.filter((c) => {
            const daysSince = Math.floor((Date.now() - new Date(c.created_at ?? '').getTime()) / (1000 * 60 * 60 * 24));
            return daysSince <= 30;
          }).length],
          ['Atelier tier', customers.filter((c) => (c.lifetime_value_ghs_minor ?? 0) > 1000000).length],
          ['Churn risk', customers.filter((c) => {
            if (!c.last_order_at) return false;
            const daysSinceLast = Math.floor((Date.now() - new Date(c.last_order_at).getTime()) / (1000 * 60 * 60 * 24));
            return daysSinceLast > 90;
          }).length],
        ].map(([label, value]) => (
          <div key={label} className="bg-white border border-line rounded-xl p-5">
            <div className="text-[10px] uppercase tracking-wider text-ink-muted">{label}</div>
            <div className="font-display text-[32px] font-normal tracking-tight mt-2">{value}</div>
          </div>
        ))}
      </div>

      {/* Customers Table */}
      <Panel>
        {/* Filter Bar */}
        <div className="flex gap-2 px-6 py-3 border-b border-line bg-[#FAFAFA] flex-wrap items-center">
          <input
            type="search"
            placeholder="Search name or email…"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="px-3 py-2 border border-line rounded-lg text-sm bg-white flex-1 min-w-[200px] focus:outline-none focus:border-lavender-400"
          />
          <select
            value={tierFilter}
            onChange={(e) => setTierFilter(e.target.value)}
            className="px-3 py-2 border border-line rounded-lg text-sm bg-white focus:outline-none focus:border-lavender-400"
          >
            <option value="">All tiers</option>
            <option value="atelier">Atelier (GH₵10k+)</option>
            <option value="bloom">Bloom (GH₵1k-10k)</option>
            <option value="petal">Petal (&lt; GH₵1k)</option>
          </select>
        </div>

        <table className="w-full border-collapse text-[13px]">
          <thead>
            <tr className="text-left">
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Customer
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Email
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Tier
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Orders
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Lifetime
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Member since
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold" />
            </tr>
          </thead>
          <tbody>
            {filteredCustomers.map((customer) => {
              const lifetime = customer.lifetime_value_ghs_minor ?? 0;

              const tier =
                lifetime > 1000000
                  ? 'Atelier'
                  : lifetime > 100000
                    ? 'Bloom'
                    : 'Petal';
              return (
                <tr key={customer.id} className="hover:bg-[#FAFAFA] transition-colors">
                  <td className="px-3 py-3 border-b border-line-soft">
                    <div className="font-semibold">{customer.name || 'No name'}</div>
                  </td>
                  <td className="px-3 py-3 border-b border-line-soft text-sm">{customer.email}</td>
                  <td className="px-3 py-3 border-b border-line-soft">
                    <span
                      className="inline-flex px-2 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider"
                      style={{
                        background: tier === 'Atelier' ? 'var(--ink)' : 'var(--lavender-100)',
                        color: tier === 'Atelier' ? 'white' : 'var(--lavender-700)',
                      }}
                    >
                      {tier}
                    </span>
                  </td>
                  <td className="px-3 py-3 border-b border-line-soft font-variant-numeric tabular-nums font-semibold">
                    {customer.order_count}
                  </td>
                  <td className="px-3 py-3 border-b border-line-soft font-variant-numeric tabular-nums font-semibold">
                    {formatCurrency(customer.lifetime_value_ghs_minor ?? 0)}
                  </td>
                  <td className="px-3 py-3 border-b border-line-soft text-sm">{formatDate(customer.created_at)}</td>
                  <td className="px-3 py-3 border-b border-line-soft">
                    <button
                      onClick={() => navigate({ to: `/admin/customers/${customer.id}` })}
                      className="text-lavender-700 font-semibold px-2 py-1 text-sm hover:underline"
                    >
                      View
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-center gap-2 mt-4">
            <button
              onClick={() => setPage(Math.max(0, page - 1))}
              disabled={page === 0}
              className="px-3 py-2 border border-line rounded-lg text-sm font-semibold hover:bg-lavender-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <span className="px-4 py-2 text-sm text-ink-muted">Page {page + 1} of {totalPages}</span>
            <button
              onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
              disabled={page >= totalPages - 1}
              className="px-3 py-2 border border-line rounded-lg text-sm font-semibold hover:bg-lavender-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        )}
      </Panel>
    </div>
  );
}
