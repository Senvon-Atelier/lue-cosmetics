import { createRouter, createRoute, createRootRoute, Outlet, Link } from '@tanstack/react-router';
import { QueryProvider } from './features/shared/providers/query-provider';
import { AuthProvider, useAuth } from './lib/auth/auth-provider';
import { CartProvider } from './features/cart/cart-provider';
import { Brand, Button } from './features/shared/ui';

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
  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      <header className="border-b border-line-soft">
        <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '1rem 2rem' }}>
          <div className="flex items-center justify-between">
            <Link to={HomeRoute.to}>
              <Brand />
            </Link>
            <nav className="flex gap-6">
              <Link to={HomeRoute.to} className="[&.active]:font-semibold">
                Home
              </Link>
              <Link to={ShopRoute.to} className="[&.active]:font-semibold">
                Shop
              </Link>
              <AuthLink />
            </nav>
          </div>
        </div>
      </header>
      <Outlet />
      <footer className="border-t border-line-soft mt-16">
        <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
          <div className="text-center text-ink-muted text-sm">
            © 2026 Rue Cosmetics Ghana · All rights reserved
          </div>
        </div>
      </footer>
    </div>
  );
}

function AuthLink() {
  const { isAuthenticated, isAdmin } = useAuth();

  if (isAdmin) {
    return (
      <Link to={AdminRoute.to} className="[&.active]:font-semibold">
        Admin
      </Link>
    );
  }

  if (isAuthenticated) {
    return (
      <Link to={AccountRoute.to} className="[&.active]:font-semibold">
        Account
      </Link>
    );
  }

  return (
    <Link to={LoginRoute.to} className="[&.active]:font-semibold">
      Login
    </Link>
  );
}

// Home route
const HomeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: () => (
    <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
      <h1 className="font-display text-4xl mb-4">Rue Cosmetics</h1>
      <p className="text-ink-muted mb-6">
        Frontend foundation is complete. Phase 4 (Router Setup) is working.
      </p>

      <div className="space-y-4">
        <div className="p-4 bg-lavender-50 rounded">
          <h2 className="font-label font-semibold mb-2">✅ What's Working</h2>
          <ul className="space-y-1 text-ink-soft">
            <li>• Project scaffolding with Vite + React 18 + TypeScript</li>
            <li>• Tailwind CSS v4 with Rue design tokens</li>
            <li>• Shared UI components (Icon, Brand, Button, Placeholder)</li>
            <li>• Orval-generated API client from backend OpenAPI spec</li>
            <li>• Auth provider with session management</li>
            <li>• Cart provider with guest/auth merge logic</li>
            <li>• TanStack Query for data fetching</li>
            <li>• TanStack Router with manual route setup</li>
          </ul>
        </div>

        <div className="p-4 bg-lavender-100 rounded">
          <h2 className="font-label font-semibold mb-2">🚧 Next Steps</h2>
          <ul className="space-y-1 text-ink-soft">
            <li>• Phase 5: Product catalog features (ProductCard, ShopPage)</li>
            <li>• Phase 6: Cart & checkout flow</li>
            <li>• Phase 7: Home page & marketing content</li>
            <li>• Phase 8: Testing & deployment</li>
          </ul>
        </div>

        <div className="flex gap-2">
          <Link to={ShopRoute.to}>
            <Button variant="primary">Shop Now</Button>
          </Link>
          <Button variant="outline">Learn More</Button>
        </div>
      </div>
    </div>
  ),
});

// Shop route
const ShopRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/shop',
  component: () => (
    <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
      <h1 className="font-display text-4xl mb-4">Shop</h1>
      <p className="text-ink-muted">Browse our curated collection of skincare, haircare, and wellness products.</p>
      <p className="mt-4 text-sm text-ink-muted">Product catalog coming in Phase 5...</p>
    </div>
  ),
});

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

// Create route tree
const routeTree = rootRoute.addChildren([
  HomeRoute,
  ShopRoute,
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
