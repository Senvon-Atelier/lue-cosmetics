import { Link } from '@tanstack/react-router';
import { Icon } from '../ui/icons';
import { Brand } from '../ui/brand';
import { useAuth } from '../../../lib/auth/auth-provider';
import { useCart } from '../../cart/cart-provider';

export function Header() {
  const { isAuthenticated } = useAuth();
  const { itemCount, wishlistCount, openDrawer } = useCart();

  return (
    <header className="header">
      <div className="wrap header-inner">
          <nav className="header-nav">
            <Link to="/" className="header-nav-link">
              Home
            </Link>
            <Link to="/shop" className="header-nav-link">
              Shop
            </Link>
            <Link to="/about" className="header-nav-link">
              About
            </Link>
            <Link to="/" hash="journal" className="header-nav-link">
              Journal
            </Link>
          </nav>
          <Link to="/">
            <Brand />
          </Link>
          <div className="header-actions">
            <button className="header-icon-btn" aria-label="Search">
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
            <button
              className="header-icon-btn"
              style={{ position: 'relative' }}
              aria-label="Wishlist"
            >
              <Icon name="heart" size={20} />
              {wishlistCount > 0 && (
                <span className="badge">
                  {wishlistCount > 9 ? '9+' : wishlistCount}
                </span>
              )}
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
            >
              <Icon name="menu" size={20} />
            </button>
          </div>
      </div>
    </header>
  );
}
