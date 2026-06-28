import { useState } from 'react';
import { createRouter, createRoute, createRootRoute, Outlet, Link, useParams } from '@tanstack/react-router';
import { QueryProvider } from './features/shared/providers/query-provider';
import { AuthProvider, useAuth } from './lib/auth/auth-provider';
import { CartProvider, useCart } from './features/cart/cart-provider';
import { Brand, Button, Icon } from './features/shared/ui';
import { ShopPage } from './features/catalog/shop-page';
import { ProductDetail } from './features/catalog/product-detail';
import { CartDrawer } from './features/cart/cart-drawer';
import { CartPage } from './features/cart/cart-page';
import { CheckoutPage } from './features/checkout/checkout-page';
import { CheckoutReturnPage } from './features/checkout/checkout-return';
import { HomeHero } from './features/home/home-hero';
import { CategoryRail } from './features/home/category-rail';
import { FeaturedProducts } from './features/home/featured-products';
import { PromiseSection } from './features/home/promise-section';
import { JournalSection } from './features/home/journal-section';
import { TestimonialsSection } from './features/home/testimonials-section';
import { NewsletterSection } from './features/home/newsletter-section';
import { AboutPage } from './features/content/about-page';

// Root route with all providers
const rootRoute = createRootRoute({
  component: () => (
    <QueryProvider>
      <AuthProvider>
        <CartProvider>
          <RootLayout />
        </CartProvider>
      </AuthProvider>
    </QueryProvider>
  ),
});

function RootLayout() {
  const [isCartOpen, setIsCartOpen] = useState(false);

  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      <header className="header">
        <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '0 2rem' }}>
          <div className="header-inner">
            <nav className="header-nav">
              <Link to={HomeRoute.to} className="header-nav-link">
                Home
              </Link>
              <Link to={ShopRoute.to} className="header-nav-link">
                Shop
              </Link>
              <Link to={AboutRoute.to} className="header-nav-link">
                About
              </Link>
            </nav>
            <Link to={HomeRoute.to}>
              <Brand />
            </Link>
            <div className="header-actions">
              <button className="header-icon-btn" aria-label="Search">
                <Icon name="search" size={20} />
              </button>
              <AuthLinkHeader />
              <CartTrigger onClick={() => setIsCartOpen(true)} />
            </div>
          </div>
        </div>
      </header>
      <Outlet />
      <footer className="footer">
        <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '0 2rem' }}>
          <div className="footer-inner">
            <div className="footer-brand">
              <div className="footer-brand-logo">
                <Brand />
              </div>
              <p className="footer-blurb">
                Home of authentic beauty and wellness. A shelf of trusted names — and a few of our own — stocked in Accra, shipped across Ghana.
              </p>
              <div className="footer-socials">
                <a href="#" className="footer-social-link" aria-label="Instagram">
                  <Icon name="instagram" size={18} />
                </a>
                <a href="#" className="footer-social-link" aria-label="TikTok">
                  <Icon name="tiktok" size={18} />
                </a>
                <a href="#" className="footer-social-link" aria-label="WhatsApp">
                  <Icon name="whatsapp" size={18} />
                </a>
              </div>
            </div>
            <div className="footer-cols">
              <div className="footer-col">
                <h5>Shop</h5>
                <ul>
                  <li><Link to="/shop">All Products</Link></li>
                  <li><Link to="/shop">Skincare</Link></li>
                  <li><Link to="/shop">Haircare</Link></li>
                  <li><Link to="/shop">Wellness</Link></li>
                </ul>
              </div>
              <div className="footer-col">
                <h5>Company</h5>
                <ul>
                  <li><Link to="/about">About Us</Link></li>
                  <li><Link to="/about">Our Story</Link></li>
                  <li><Link to="/about">Careers</Link></li>
                </ul>
              </div>
              <div className="footer-col">
                <h5>Help</h5>
                <ul>
                  <li><Link to="/account">My Account</Link></li>
                  <li><Link to="/account">Order Status</Link></li>
                  <li><Link to="/about">Contact Us</Link></li>
                </ul>
              </div>
              <div className="footer-col">
                <h5>Visit</h5>
                <ul>
                  <li>Spintex Road, Accra</li>
                  <li>+233 20 123 4567</li>
                  <li>Mon-Sat: 10am-7pm</li>
                </ul>
              </div>
            </div>
          </div>
          <div className="footer-bottom">
            <p>© 2026 Rue Cosmetics Ghana · All rights reserved</p>
            <div className="footer-legal">
              <a href="/legal/privacy">Privacy</a>
              <a href="/legal/terms">Terms</a>
              <a href="/legal/returns">Returns</a>
            </div>
          </div>
        </div>
      </footer>
      <CartDrawer open={isCartOpen} onClose={() => setIsCartOpen(false)} />
    </div>
  );
}

function AuthLinkHeader() {
  const { isAuthenticated } = useAuth();

  if (isAuthenticated) {
    return (
      <Link to={AccountRoute.to} className="header-icon-btn" aria-label="Account">
        <Icon name="user" size={20} />
      </Link>
    );
  }

  return (
    <Link to={LoginRoute.to} className="header-icon-btn" aria-label="Account">
      <Icon name="user" size={20} />
    </Link>
  );
}

function CartTrigger({ onClick }: { onClick: () => void }) {
  const { itemCount } = useCart();

  return (
    <button
      onClick={onClick}
      className="relative p-2 rounded-full hover:bg-lavender-50 transition-colors"
      aria-label="Open cart"
    >
      <Icon name="bag" size={20} />
      {itemCount > 0 && (
        <span className="absolute -top-1 -right-1 w-5 h-5 bg-lavender-600 text-paper text-xs font-label font-medium rounded-full flex items-center justify-center">
          {itemCount > 9 ? '9+' : itemCount}
        </span>
      )}
    </button>
  );
}

// Home route
const HomeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: () => (
    <div className="min-h-screen bg-paper text-ink font-body">
      <HomeHero />
      <PromiseSection />
      <CategoryRail />
      <FeaturedProducts />
      <JournalSection />
      <TestimonialsSection />
      <NewsletterSection />
    </div>
  ),
});

// Shop route
const ShopRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/shop',
  component: () => (
    <div className="min-h-screen bg-paper text-ink font-body">
      <div className="mb-6 pt-6" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem 2rem 0' }}>
        <h1 className="font-display text-4xl mb-2">Shop</h1>
        <p className="text-ink-muted">Browse our curated collection of skincare, haircare, and wellness products.</p>
      </div>
      <ShopPage />
    </div>
  ),
});

// Product detail route
const ProductDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/shop/$slug',
  component: ProductDetailComponent,
});

function ProductDetailComponent() {
  const { slug } = useParams({ from: '/shop/$slug' });
  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      <ProductDetail slug={slug || ''} />
    </div>
  );
}

// Login route
const LoginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: () => (
    <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
      <h1 className="font-display text-4xl mb-4">Login</h1>
      <p className="text-ink-muted mb-6">Sign in to your account to access your wishlist and order history.</p>
      <p className="text-sm text-ink-muted">Login form coming soon...</p>
    </div>
  ),
});

// Account route (protected)
const AccountRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/account',
  component: AccountComponent,
});

function AccountComponent() {
  const { user, isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-ink-muted">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
        <h1 className="font-display text-4xl mb-4">Account Required</h1>
        <p className="text-ink-muted">Please log in to access your account.</p>
      </div>
    );
  }

  return (
    <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
      <h1 className="font-display text-4xl mb-4">Welcome back, {user?.name || user?.email}!</h1>
      <p className="text-ink-muted mb-6">
        {user?.role === 'admin' ? 'You have admin access.' : 'Manage your orders, addresses, and wishlist.'}
      </p>

      <div className="space-y-4">
        <div className="p-4 bg-lavender-50 rounded">
          <h2 className="font-label font-semibold mb-2">Account Details</h2>
          <ul className="space-y-1 text-ink-soft">
            <li>• Email: {user?.email}</li>
            <li>• Email verified: {user?.email_verified ? 'Yes' : 'No'}</li>
            <li>• Role: {user?.role}</li>
            <li>• User ID: {user?.user_id}</li>
          </ul>
        </div>

        <p className="text-sm text-ink-muted">More account features coming soon...</p>
      </div>
    </div>
  );
}

// Admin route (protected + admin only)
const AdminRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/admin',
  component: AdminComponent,
});

function AdminComponent() {
  const { user, isAuthenticated, isAdmin, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-ink-muted">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
        <h1 className="font-display text-4xl mb-4">Access Denied</h1>
        <p className="text-ink-muted">Admin access required. Please log in as an administrator.</p>
      </div>
    );
  }

  if (!isAdmin) {
    return (
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
        <h1 className="font-display text-4xl mb-4">Access Denied</h1>
        <p className="text-ink-muted">You don't have permission to access the admin dashboard.</p>
      </div>
    );
  }

  return (
    <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
      <h1 className="font-display text-4xl mb-4">Admin Dashboard</h1>
      <p className="text-ink-muted mb-6">
        Welcome, {user?.name || user?.email}. You have admin access.
      </p>

      <div className="space-y-4">
        <div className="p-4 bg-lavender-50 rounded">
          <h2 className="font-label font-semibold mb-2">Admin Features</h2>
          <ul className="space-y-1 text-ink-soft">
            <li>• View all orders</li>
            <li>• Update order status</li>
            <li>• Manage users</li>
            <li>• View revenue stats</li>
          </ul>
        </div>

        <p className="text-sm text-ink-muted">Admin features coming soon...</p>
      </div>
    </div>
  );
}

// Cart route
const CartRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/cart',
  component: () => (
    <div className="min-h-screen bg-paper text-ink font-body">
      <CartPage />
    </div>
  ),
});

// Checkout route (protected)
const CheckoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/checkout',
  component: CheckoutComponent,
});

function CheckoutComponent() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-ink-muted">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
        <h1 className="font-display text-4xl mb-4">Login Required</h1>
        <p className="text-ink-muted mb-6">Please log in to proceed with checkout.</p>
        <Link to="/login">
          <Button variant="primary">Go to Login</Button>
        </Link>
      </div>
    );
  }

  return <CheckoutPage />;
}

// Checkout return route
const CheckoutReturnRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/checkout/return',
  component: () => (
    <div className="min-h-screen bg-paper text-ink font-body">
      <CheckoutReturnPage />
    </div>
  ),
});

// About route
const AboutRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/about',
  component: () => (
    <div className="min-h-screen bg-paper text-ink font-body">
      <AboutPage />
    </div>
  ),
});

// Create route tree
const routeTree = rootRoute.addChildren([
  HomeRoute,
  ShopRoute,
  ProductDetailRoute,
  CartRoute,
  CheckoutRoute,
  CheckoutReturnRoute,
  AboutRoute,
  LoginRoute,
  AccountRoute,
  AdminRoute,
]);

// Create the router with route tree
export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
  defaultPreloadStaleTime: 0,
});

// Register router for TypeScript inference
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
