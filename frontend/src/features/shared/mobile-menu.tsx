import { useEffect, useState } from 'react';
import { Link } from '@tanstack/react-router';
import { getCategories } from '../../lib/api/generated/rueCosmeticsAPI';
import { STORE_INFO } from '../../content/store-info';
import { Brand } from './ui/brand';
import { Icon } from './ui/icons';
import { useEscToClose, useLockBodyScroll } from './use-overlay';

// Ported from Rue/src/shared.jsx MobileMenu (lines 336–382).
// No category counts (.mnav-count) — the API has none. Contact copy shared with the footer.

interface MobileMenuProps {
  open: boolean;
  onClose: () => void;
}

type CategoryLink = { id?: string; label?: string; slug?: string };

export function MobileMenu({ open, onClose }: MobileMenuProps) {
  const [categories, setCategories] = useState<CategoryLink[]>([]);

  useEscToClose(open, onClose);
  useLockBodyScroll(open);

  // Fetch categories the first time the menu opens (not on every page load)
  useEffect(() => {
    if (!open || categories.length > 0) return;
    getCategories()
      .then((cats) => setCategories(cats ?? []))
      .catch(() => {
        /* categories section simply stays empty — nav links still work */
      });
  }, [open, categories.length]);

  return (
    <>
      <div className={`drawer-scrim${open ? ' open' : ''}`} onClick={onClose} />
      <aside className={`drawer drawer-left${open ? ' open' : ''}`} inert={open ? undefined : ''} aria-label="Menu">
        <div className="drawer-head">
          <Link to="/" onClick={onClose} aria-label="Rue home">
            <Brand />
          </Link>
          <button className="icon-btn" onClick={onClose} aria-label="Close menu">
            <Icon name="close" />
          </button>
        </div>
        <div className="drawer-body mobile-nav">
          <div className="mnav-section">
            <div className="eyebrow">Pages</div>
            <Link to="/" onClick={onClose} activeOptions={{ exact: true }} activeProps={{ className: 'active' }}>
              Home <Icon name="chevronRight" size={14} />
            </Link>
            <Link to="/shop" onClick={onClose} activeProps={{ className: 'active' }}>
              Shop <Icon name="chevronRight" size={14} />
            </Link>
            <Link to="/about" onClick={onClose} activeProps={{ className: 'active' }}>
              About <Icon name="chevronRight" size={14} />
            </Link>
            <Link to="/" hash="journal" onClick={onClose}>
              Journal <Icon name="chevronRight" size={14} />
            </Link>
          </div>
          {categories.length > 0 && (
            <div className="mnav-section">
              <div className="eyebrow">Shop by category</div>
              {categories.map((c) => (
                <Link key={c.id} to="/shop" search={{ category: c.slug }} onClick={onClose}>
                  {c.label} <Icon name="chevronRight" size={14} />
                </Link>
              ))}
            </div>
          )}
          <div className="mnav-section mnav-contact">
            <div className="eyebrow">Visit us</div>
            <p>
              <Icon name="pin" size={14} /> {STORE_INFO.addressLine1} · {STORE_INFO.addressLine2}
            </p>
            <p>
              <Icon name="phone" size={14} /> {STORE_INFO.phone}
            </p>
            <p>
              <Icon name="clock" size={14} /> {STORE_INFO.hours}
            </p>
          </div>
        </div>
      </aside>
    </>
  );
}
