import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useGetAdminCustomers } from '../../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs, formatOrderDate } from '../../../lib/format/utils';
import { Panel } from '../../shared/ui/admin';

export function AdminCustomers() {
  const navigate = useNavigate();
  const [page, setPage] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');

  const { data: customersData, isLoading, error } = useGetAdminCustomers({
    page: page + 1,
    page_size: 20,
  });

  const totalPages = customersData?.total_pages ?? 1;
  const customers = customersData?.customers ?? [];

  // Client-side search over the fetched page (pre-existing; server-side search is a backlog item)
  const filteredCustomers = customers.filter((c) => {
    if (
      searchQuery &&
      !c.name?.toLowerCase().includes(searchQuery.toLowerCase()) &&
      !(c.email ?? '').toLowerCase().includes(searchQuery.toLowerCase())
    ) {
      return false;
    }
    return true;
  });

  if (isLoading) {
    return <div className="admin-loading">Loading customers…</div>;
  }

  if (error) {
    return (
      <div className="alert alert-warn">
        Failed to load customers: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">CRM</div>
          <h1>Customers</h1>
        </div>
      </div>

      <Panel flush>
        <div className="admin-filter-bar">
          <input
            type="search"
            placeholder="Search name or email…"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        {filteredCustomers.length === 0 ? (
          <div className="admin-panel-body">
            <p className="admin-empty">No customers match.</p>
          </div>
        ) : (
          <table className="admin-tbl">
            <thead>
              <tr>
                <th>Customer</th>
                <th>Email</th>
                <th>Orders</th>
                <th>Lifetime</th>
                <th>Member since</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {filteredCustomers.map((customer) => (
                <tr key={customer.id}>
                  <td>{customer.name || 'No name'}</td>
                  <td>{customer.email}</td>
                  <td className="num">{customer.order_count}</td>
                  <td className="num">{formatGhs(customer.lifetime_value_ghs_minor ?? 0)}</td>
                  <td>{formatOrderDate(customer.created_at)}</td>
                  <td>
                    <button
                      className="admin-btn admin-btn-link"
                      onClick={() => navigate({ to: `/admin/customers/${customer.id}` })}
                    >
                      View
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {totalPages > 1 && (
          <div className="admin-pagination">
            <button
              className="admin-btn admin-btn-sec"
              onClick={() => setPage(Math.max(0, page - 1))}
              disabled={page === 0}
            >
              Previous
            </button>
            <span>
              Page {page + 1} of {totalPages}
            </span>
            <button
              className="admin-btn admin-btn-sec"
              onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
              disabled={page >= totalPages - 1}
            >
              Next
            </button>
          </div>
        )}
      </Panel>
    </>
  );
}
