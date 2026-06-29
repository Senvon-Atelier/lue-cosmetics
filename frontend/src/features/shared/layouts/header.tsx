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
  const { itemCount } = useCart();

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
              onClick={onCartOpen}
              aria-label="Open cart"
            >
              <Icon name="bag" size={20} />
              {itemCount > 0 && (
                <span className="absolute -top-1 -right-1 w-5 h-5 bg-lavender-600 text-paper text-xs font-label font-medium rounded-full flex items-center justify-center">
                  {itemCount > 9 ? '9+' : itemCount}
                </span>
              )}
            </button>
          </div>
        </div>
      </div>
    </header>
  );
}
