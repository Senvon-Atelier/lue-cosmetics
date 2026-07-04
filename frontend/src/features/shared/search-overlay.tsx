import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { useGetProducts } from '../../lib/api/generated/rueCosmeticsAPI';
import type { InternalCatalogProductView } from '../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs, getImageUrl } from '../../lib/format/utils';
import { SEARCH_TERMS } from '../../content/search-terms';
import { useDebouncedValue } from '../../lib/hooks/use-debounced-value';
import { Icon } from './ui/icons';
import { useEscToClose, useLockBodyScroll } from './use-overlay';

// Ported from Rue/src/shared.jsx SearchOverlay (lines 244–335); real data via GET /products?q=.
// Idle sections: curated trending chips + honest "From the shop" rail (no fake popularity).

interface SearchOverlayProps {
  open: boolean;
  onClose: () => void;
}

export function SearchOverlay({ open, onClose }: SearchOverlayProps) {
  const [q, setQ] = useState('');
  const debouncedQ = useDebouncedValue(q, 300);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  useEscToClose(open, onClose);
  useLockBodyScroll(open);

  // Reset query when the overlay closes so it doesn't persist on reopen
  useEffect(() => {
    if (!open) setQ('');
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const t = setTimeout(() => inputRef.current?.focus(), 100);
    return () => clearTimeout(t);
  }, [open]);

  const term = debouncedQ.trim();

  const { data: resultsData, isLoading: searching, error: searchError } = useGetProducts(
    { q: term, limit: 6 },
    { query: { enabled: open && term.length > 0 } },
  );
  const { data: idleData } = useGetProducts(
    { limit: 3 },
    { query: { enabled: open && term.length === 0 } },
  );
  const results = resultsData?.items ?? [];
  const idlePicks = idleData?.items ?? [];

  const openProduct = (slug?: string) => {
    if (!slug) return;
    onClose();
    setQ('');
    void navigate({ to: `/shop/${slug}` });
  };

  return (
    <div className={`search-overlay${open ? ' open' : ''}`} inert={open ? undefined : ''}>
      <div className="search-head wrap">
        <div className="search-input-wrap">
          <Icon name="search" size={20} />
          <input
            ref={inputRef}
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Search products, brands, rituals..."
            aria-label="Search products"
            className="search-input"
          />
          {q && (
            <button onClick={() => setQ('')} className="search-clear" aria-label="Clear search">
              <Icon name="close" size={14} />
            </button>
          )}
        </div>
        <button className="icon-btn" onClick={onClose} aria-label="Close search">
          <Icon name="close" />
        </button>
      </div>
      <div className="search-body wrap">
        {term.length === 0 ? (
          <>
            <div className="search-section">
              <div className="eyebrow">Trending searches</div>
              <div className="search-chips">
                {SEARCH_TERMS.map((t) => (
                  <button key={t} className="chip" onClick={() => setQ(t)}>
                    {t}
                  </button>
                ))}
              </div>
            </div>
            {idlePicks.length > 0 && (
              <div className="search-section">
                <div className="eyebrow">From the shop</div>
                <div className="search-picks">
                  {idlePicks.map((p) => (
                    <SearchPick key={p.id} product={p} onOpen={openProduct} />
                  ))}
                </div>
              </div>
            )}
          </>
        ) : searchError ? (
          <div className="search-empty">
            <p>Search is unavailable right now. Please try again in a moment.</p>
          </div>
        ) : searching ? (
          <div className="search-empty">
            <p>Searching…</p>
          </div>
        ) : results.length === 0 ? (
          <div className="search-empty">
            <p>
              No results for <em>"{term}"</em>. Try a different term or browse{' '}
              <Link to="/shop" search={{}} onClick={onClose}>
                the full shop
              </Link>
              .
            </p>
          </div>
        ) : (
          <div className="search-section">
            <div className="eyebrow">
              {results.length} result{results.length === 1 ? '' : 's'}
            </div>
            <div className="search-picks">
              {results.map((p) => (
                <SearchPick key={p.id} product={p} onOpen={openProduct} />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function SearchPick({
  product,
  onOpen,
}: {
  product: InternalCatalogProductView;
  onOpen: (slug?: string) => void;
}) {
  return (
    <button type="button" className="search-pick" onClick={() => onOpen(product.slug)}>
      {product.image_path ? (
        <img
          src={getImageUrl(product.image_path)}
          alt=""
          style={{ width: 64, height: 80, objectFit: 'cover', borderRadius: 'var(--radius)', flexShrink: 0 }}
          loading="lazy"
        />
      ) : (
        <div className={`ph ph--${product.tone ?? 'lavender'}`} style={{ width: 64, height: 80, flexShrink: 0 }}>
          <span className="ph-label" style={{ fontSize: 8 }}>
            {product.name?.slice(0, 2)}
          </span>
        </div>
      )}
      <div>
        <div className="cart-item-name">{product.name}</div>
        <div className="price" style={{ marginTop: 4 }}>
          {formatGhs(product.price_ghs_minor ?? 0)}
        </div>
      </div>
    </button>
  );
}
