import { Link } from '@tanstack/react-router';
import { Icon } from '../ui/icons';
import { Brand } from '../ui/brand';
import { useAuth } from '../../../lib/auth/auth-provider';
import { useCart } from '../../cart/cart-provider';

interface HeaderProps {
  onCartOpen: () => void;
}

export function Header({ onCartOpen }: HeaderProps) {
  const { isAuthenticated } = useAuth();
  const { itemCount, wishlistCount } = useCart();

  return (
    <header className="header">
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '0 2rem' }}>
        <div className="header-inner">
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
            <Link to="/journal" className="header-nav-link">
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
              className="header-icon-btn relative"
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
              className="header-icon-btn relative"
              onClick={onCartOpen}
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
      </div>
    </header>
  );
}
