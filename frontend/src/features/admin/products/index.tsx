import { useState } from 'react';
import { useGetAdminProducts } from '../../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs, getImageUrl } from '../../../lib/format/utils';
import { KPICard, Panel } from '../../shared/ui/admin';

export function AdminProducts() {
  const [page, setPage] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');

  const { data: productsData, isLoading, error } = useGetAdminProducts({
    page: page + 1,
    page_size: 20,
  });

  const totalPages = productsData?.total_pages ?? 1;
  const products = productsData?.products ?? [];

  // Client-side search over the fetched page (pre-existing; server-side search is a backlog item)
  const filteredProducts = products.filter((p) => {
    if (
      searchQuery &&
      !(p.name ?? '').toLowerCase().includes(searchQuery.toLowerCase()) &&
      !(p.slug ?? '').toLowerCase().includes(searchQuery.toLowerCase())
    ) {
      return false;
    }
    return true;
  });

  if (isLoading) {
    return <div className="admin-loading">Loading products…</div>;
  }

  if (error) {
    return (
      <div className="alert alert-warn">
        Failed to load products: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">Catalog</div>
          <h1>Products</h1>
        </div>
      </div>

      <div className="admin-kpis">
        <KPICard title="Total SKUs" value={productsData?.total ?? '—'} />
      </div>

      <Panel flush>
        <div className="admin-filter-bar">
          <input
            type="search"
            placeholder="Search name or slug…"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        {filteredProducts.length === 0 ? (
          <div className="admin-panel-body">
            <p className="admin-empty">No products match.</p>
          </div>
        ) : (
          <table className="admin-tbl">
            <thead>
              <tr>
                <th>Product</th>
                <th>Category</th>
                <th>Brand</th>
                <th>Price</th>
              </tr>
            </thead>
            <tbody>
              {filteredProducts.map((product) => (
                <tr key={product.id}>
                  <td>
                    <div className="row-prod">
                      {product.image_path ? (
                        <img
                          src={getImageUrl(product.image_path)}
                          alt={product.name}
                          loading="lazy"
                        />
                      ) : (
                        <div className="ph ph--lavender ph-sm" />
                      )}
                      <div>
                        <div className="row-prod-name">{product.name}</div>
                        <div className="row-prod-sku">{product.slug}</div>
                      </div>
                    </div>
                  </td>
                  <td>{(product.category_id ?? '—').slice(0, 8)}</td>
                  <td>{(product.brand_id ?? '—').slice(0, 8)}</td>
                  <td className="num">{formatGhs(product.price_ghs_minor ?? 0)}</td>
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
