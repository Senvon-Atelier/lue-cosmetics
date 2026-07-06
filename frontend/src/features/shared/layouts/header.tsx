import { useState } from 'react';
import { Link } from '@tanstack/react-router';
import { Icon } from '../ui/icons';
import { Brand } from '../ui/brand';
import { useAuth } from '../../../lib/auth/auth-provider';
import { useCart } from '../../cart/cart-provider';
import { SearchOverlay } from '../search-overlay';
import { MobileMenu } from '../mobile-menu';

export function Header() {
  const { isAuthenticated } = useAuth();
  const { itemCount, openDrawer } = useCart();
  const [searchOpen, setSearchOpen] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <>
    <header className="header">
      <div className="wrap header-inner">
          <nav className="header-nav">
            <Link to="/" className="header-nav-link">
              Home
            </Link>
            <Link to="/shop" search={{}} className="header-nav-link">
              Shop
            </Link>
            <Link to="/about" className="header-nav-link">
              About
            </Link>
            <Link to="/" hash="journal" className="header-nav-link">
              Journal
            </Link>
          </nav>
          <Link to="/" className="brand">
            <Brand />
          </Link>
          <div className="header-actions">
            <button className="header-icon-btn" aria-label="Search" onClick={() => setSearchOpen(true)}>
              <Icon name="search" size={20} />
            </button>
            {isAuthenticated ? (
              <Link to="/account" className="header-icon-btn" aria-label="Account">
                <Icon name="user" size={20} />
              </Link>
            ) : (
              <Link to="/login" className="header-icon-btn" aria-label="Account">
                <Icon name="user" size={20} />
              </Link>
            )}
            <button className="header-icon-btn" aria-label="Wishlist" disabled title="Saved items coming soon">
              <Icon name="heart" size={20} />
            </button>
            <button
              className="header-icon-btn"
              style={{ position: 'relative' }}
              onClick={openDrawer}
              aria-label="Open cart"
            >
              <Icon name="bag" size={20} />
              {itemCount > 0 && (
                <span className="badge">
                  {itemCount > 9 ? '9+' : itemCount}
                </span>
              )}
            </button>
            <button
              className="header-icon-btn mobile-menu-btn"
              aria-label="Menu"
              onClick={() => setMenuOpen(true)}
            >
              <Icon name="menu" size={20} />
            </button>
          </div>
      </div>
    </header>
    <SearchOverlay open={searchOpen} onClose={() => setSearchOpen(false)} />
    <MobileMenu open={menuOpen} onClose={() => setMenuOpen(false)} />
    </>);
}
