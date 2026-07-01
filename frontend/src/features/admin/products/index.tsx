import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { getApiV1AdminProducts } from '../../../lib/api/generated/rueCosmeticsAPI';
import { StatusTag, Panel } from '../../shared/ui/admin';

export function AdminProducts() {
  const [page, setPage] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  const { data: productsData, isLoading, error } = useQuery({
    queryKey: ['admin', 'products', page, searchQuery, categoryFilter, statusFilter],
    queryFn: () =>
      getApiV1AdminProducts({
        page: page + 1,
        page_size: 20,
      }),
  });

  const formatCurrency = (amount: number) => {
    return `GH₵${(amount / 100).toLocaleString()}`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const getStockStatus = (product: any) => {
    // This is a placeholder - you'll need to add stock tracking to the product model
    return 'live';
  };

  const totalPages = productsData?.data.total_pages || 1;
  const products = productsData?.data.products || [];

  // Filter products based on search (client-side for now)
  const filteredProducts = products.filter((p) => {
    if (searchQuery && !p.name.toLowerCase().includes(searchQuery.toLowerCase()) && !p.slug.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false;
    }
    return true;
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-ink-muted">Loading products...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg">
        Failed to load products: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  return (
    <div>
      {/* Header */}
      <div className="flex items-start justify-between mb-7 gap-4 flex-wrap">
        <div>
          <div className="text-lavender-700 text-sm mb-1">Catalog</div>
          <h1 className="font-display text-4xl font-normal">Products</h1>
        </div>
        <div className="flex gap-2">
          <button className="px-3 py-2 border border-line rounded-lg text-sm font-semibold hover:bg-lavender-50 transition-colors">
            Import CSV
          </button>
          <button className="px-3 py-2 bg-ink text-white rounded-lg text-sm font-semibold hover:bg-lavender-700 transition-colors">
            New product
          </button>
        </div>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-4 gap-3 mb-6">
        {[
          ['Total SKUs', products.length],
          ['Live', products.filter((p) => getStockStatus(p) === 'live').length],
          ['Low stock', products.filter((p) => getStockStatus(p) === 'low').length],
          ['Out of stock', products.filter((p) => getStockStatus(p) === 'oos').length],
        ].map(([label, value]) => (
          <div key={label} className="bg-white border border-line rounded-xl p-5">
            <div className="text-[10px] uppercase tracking-wider text-ink-muted">{label}</div>
            <div className="font-display text-[32px] font-normal tracking-tight mt-2">{value}</div>
          </div>
        ))}
      </div>

      {/* Products Table */}
      <Panel>
        {/* Filter Bar */}
        <div className="flex gap-2 px-6 py-3 border-b border-line bg-[#FAFAFA] flex-wrap items-center">
          <input
            type="search"
            placeholder="Search SKU, name, brand…"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="px-3 py-2 border border-line rounded-lg text-sm bg-white flex-1 min-w-[200px] focus:outline-none focus:border-lavender-400"
          />
          <select
            value={categoryFilter}
            onChange={(e) => setCategoryFilter(e.target.value)}
            className="px-3 py-2 border border-line rounded-lg text-sm bg-white focus:outline-none focus:border-lavender-400"
          >
            <option value="">All categories</option>
            <option value="skincare">Skincare</option>
            <option value="haircare">Haircare</option>
            <option value="bodycare">Bodycare</option>
            <option value="fragrance">Fragrance</option>
          </select>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 border border-line rounded-lg text-sm bg-white focus:outline-none focus:border-lavender-400"
          >
            <option value="">All statuses</option>
            <option value="live">Live</option>
            <option value="low">Low stock</option>
            <option value="oos">Out of stock</option>
            <option value="draft">Draft</option>
          </select>
        </div>

        <table className="w-full border-collapse text-[13px]">
          <thead>
            <tr className="text-left">
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Product
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Category
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Brand ID
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Price
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Stock
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Status
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold" />
            </tr>
          </thead>
          <tbody>
            {filteredProducts.map((product) => (
              <tr key={product.id} className="hover:bg-[#FAFAFA] transition-colors">
                <td className="px-3 py-3 border-b border-line-soft">
                  <div className="flex items-center gap-3">
                    {product.image_path && (
                      <img
                        src={product.image_path}
                        alt={product.name}
                        className="w-11 h-11 rounded object-cover"
                        loading="lazy"
                      />
                    )}
                    <div>
                      <div className="font-display text-sm">{product.name}</div>
                      <div className="text-xs text-ink-muted">{product.slug}</div>
                    </div>
                  </div>
                </td>
                <td className="px-3 py-3 border-b border-line-soft capitalize">{product.category_id.slice(0, 8)}</td>
                <td className="px-3 py-3 border-b border-line-soft text-xs">{product.brand_id.slice(0, 8)}</td>
                <td className="px-3 py-3 border-b border-line-soft font-variant-numeric tabular-nums font-semibold">
                  {formatCurrency(product.price_ghs_minor)}
                </td>
                <td className="px-3 py-3 border-b border-line-soft font-variant-numeric tabular-nums">—</td>
                <td className="px-3 py-3 border-b border-line-soft">
                  <StatusTag status={getStockStatus(product)} />
                </td>
                <td className="px-3 py-3 border-b border-line-soft">
                  <button className="text-lavender-700 font-semibold px-2 py-1 text-sm hover:underline">
                    Edit
                  </button>
                </td>
              </tr>
            ))}
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
