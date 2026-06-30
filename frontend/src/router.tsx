import { createRouter, createRoute, createRootRoute, Outlet, useParams } from '@tanstack/react-router';
import { QueryProvider } from './features/shared/providers/query-provider';
import { AuthProvider } from './lib/auth/auth-provider';
import { CartProvider } from './features/cart/cart-provider';
import { RootLayout, CheckoutLayout } from './features/shared/layouts';
import { ShopPage } from './features/catalog/shop-page';
import { ProductDetail } from './features/catalog/product-detail';
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
import { useAuth } from './lib/auth/auth-provider';
import { LoginPage } from './features/auth/login-page';
import { SignupPage } from './features/auth/signup-page';
import { ForgotPasswordPage } from './features/auth/forgot-password-page';
import { ResetPasswordPage } from './features/auth/reset-password-page';
import { VerifyEmailPage } from './features/auth/verify-email-page';

// Root route with all providers (no layout)
const rootRoute = createRootRoute({
  component: () => (
    <QueryProvider>
      <AuthProvider>
        <CartProvider>
          <Outlet />
        </CartProvider>
      </AuthProvider>
    </QueryProvider>
  ),
});

// Marketing layout route (RootLayout: Header + Footer + CartDrawer)
const marketingLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: '_marketing',
  component: RootLayout,
});

// Checkout layout route (Brand + minimal chrome)
const checkoutLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: '_checkout',
  component: CheckoutLayout,
});

// Home route (child of marketing layout)
const HomeRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
  path: '/',
  component: () => (
    <div>
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

// Shop route (child of marketing layout)
const ShopRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
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

// Product detail route (child of marketing layout)
const ProductDetailRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
  path: '/shop/$slug',
  component: ProductDetailComponent,
});

function ProductDetailComponent() {
  const { slug } = useParams({ from: '/_marketing/shop/$slug' });
  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      <ProductDetail slug={slug || ''} />
    </div>
  );
}

// Cart route (child of marketing layout)
const CartRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
  path: '/cart',
  component: () => (
    <div className="min-h-screen bg-paper text-ink font-body">
      <CartPage />
    </div>
  ),
});

// Checkout routes (children of checkout layout)
const CheckoutRoute = createRoute({
  getParentRoute: () => checkoutLayoutRoute,
  path: '/checkout',
  component: CheckoutPage,
});

const CheckoutReturnRoute = createRoute({
  getParentRoute: () => checkoutLayoutRoute,
  path: '/checkout/return',
  component: () => (
    <div className="min-h-screen bg-paper text-ink font-body">
      <CheckoutReturnPage />
    </div>
  ),
});

// About route (child of marketing layout)
const AboutRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
  path: '/about',
  component: () => (
    <div className="min-h-screen bg-paper text-ink font-body">
      <AboutPage />
    </div>
  ),
});

// Login route (child of marketing layout)
const LoginRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
  path: '/login',
  component: LoginPage,
});

// Signup route (child of marketing layout)
const SignupRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
  path: '/signup',
  component: SignupPage,
});

// Forgot password route (child of marketing layout)
const ForgotPasswordRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
  path: '/forgot-password',
  component: ForgotPasswordPage,
});

// Reset password route (child of marketing layout)
const ResetPasswordRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
  path: '/reset-password',
  component: ResetPasswordPage,
});

// Verify email route (child of marketing layout)
const VerifyEmailRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
  path: '/verify-email',
  component: VerifyEmailPage,
});

// Account route (protected, child of marketing layout)
const AccountRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
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

// Admin route (protected + admin only, child of marketing layout)
const AdminRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
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

// Create route tree
const routeTree = rootRoute.addChildren([
  marketingLayoutRoute.addChildren([
    HomeRoute,
    ShopRoute,
    ProductDetailRoute,
    CartRoute,
    AboutRoute,
    LoginRoute,
    SignupRoute,
    ForgotPasswordRoute,
    ResetPasswordRoute,
    VerifyEmailRoute,
    AccountRoute,
    AdminRoute,
  ]),
  checkoutLayoutRoute.addChildren([CheckoutRoute, CheckoutReturnRoute]),
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
