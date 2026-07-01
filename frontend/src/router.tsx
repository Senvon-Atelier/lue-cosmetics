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
import { AccountLayout } from './features/account/account-layout';
import { AccountDashboard } from './features/account/account-dashboard';
import { AccountOrders } from './features/account/account-orders';
import { AccountOrderDetail } from './features/account/account-order-detail';
import { AccountAddresses } from './features/account/account-addresses';
import { AccountWishlist } from './features/account/account-wishlist';
import { AccountSettings } from './features/account/account-settings';
import { LoginPage } from './features/auth/login-page';
import { SignupPage } from './features/auth/signup-page';
import { ForgotPasswordPage } from './features/auth/forgot-password-page';
import { ResetPasswordPage } from './features/auth/reset-password-page';
import { VerifyEmailPage } from './features/auth/verify-email-page';
import { AdminLayout } from './features/admin/admin-layout';
import { AdminDashboard } from './features/admin/dashboard';
import { AdminOrders } from './features/admin/orders';
import { AdminOrderDetail } from './features/admin/orders/order-detail';
import { AdminProducts } from './features/admin/products';
import { AdminCustomers } from './features/admin/customers';
import { AdminAnalytics } from './features/admin/analytics';
import { AdminMarketing } from './features/admin/marketing';
import { AdminContent } from './features/admin/content';
import { AdminSettings } from './features/admin/settings';

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
  component: AccountLayout,
});


// Account dashboard route (child of account layout)
const AccountDashboardRoute = createRoute({
  getParentRoute: () => AccountRoute,
  path: "/",
  component: AccountDashboard,
});

// Account orders route (child of account layout)
const AccountOrdersRoute = createRoute({
  getParentRoute: () => AccountRoute,
  path: "/orders",
  component: AccountOrders,
});

// Account order detail route (child of account layout)
const AccountOrderDetailRoute = createRoute({
  getParentRoute: () => AccountOrdersRoute,
  path: "$id",
  component: AccountOrderDetail,
});

// Account addresses route (child of account layout)
const AccountAddressesRoute = createRoute({
  getParentRoute: () => AccountRoute,
  path: "/addresses",
  component: AccountAddresses,
});

// Account wishlist route (child of account layout)
const AccountWishlistRoute = createRoute({
  getParentRoute: () => AccountRoute,
  path: "/wishlist",
  component: AccountWishlist,
});

// Account settings route (child of account layout)
const AccountSettingsRoute = createRoute({
  getParentRoute: () => AccountRoute,
  path: "/settings",
  component: AccountSettings,
});


// Admin route (protected + admin only, child of marketing layout)
const AdminRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
  path: '/admin',
  component: AdminLayout,
});

const AdminDashboardRoute = createRoute({
  getParentRoute: () => AdminRoute,
  path: '/',
  component: AdminDashboard,
});

const AdminOrdersRoute = createRoute({
  getParentRoute: () => AdminRoute,
  path: '/orders',
  component: AdminOrders,
});

const AdminOrderDetailRoute = createRoute({
  getParentRoute: () => AdminRoute,
  path: '/orders/$id',
  component: AdminOrderDetail,
});

const AdminProductsRoute = createRoute({
  getParentRoute: () => AdminRoute,
  path: '/products',
  component: AdminProducts,
});

const AdminProductDetailRoute = createRoute({
  getParentRoute: () => AdminRoute,
  path: '/products/$id',
  component: () => <div>Admin Product Detail - Coming Soon...</div>,
});

const AdminCustomersRoute = createRoute({
  getParentRoute: () => AdminRoute,
  path: '/customers',
  component: AdminCustomers,
});

const AdminCustomerDetailRoute = createRoute({
  getParentRoute: () => AdminRoute,
  path: '/customers/$id',
  component: () => <div>Admin Customer Detail - Coming Soon...</div>,
});

const AdminAnalyticsRoute = createRoute({
  getParentRoute: () => AdminRoute,
  path: '/analytics',
  component: AdminAnalytics,
});

const AdminMarketingRoute = createRoute({
  getParentRoute: () => AdminRoute,
  path: '/marketing',
  component: AdminMarketing,
});

const AdminContentRoute = createRoute({
  getParentRoute: () => AdminRoute,
  path: '/content',
  component: AdminContent,
});

const AdminSettingsRoute = createRoute({
  getParentRoute: () => AdminRoute,
  path: '/settings',
  component: AdminSettings,
});


/**const AdminRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
  path: '/admin',
  component: AdminLayout,
}).addChildren([
  // Dashboard is the index
  createRoute({
    getParentRoute: () => AdminRoute,
    path: '/',
    component: AdminDashboard,
  }),
  // Orders routes
  createRoute({
    getParentRoute: () => AdminRoute,
    path: '/orders',
    component: AdminOrders,
  }),
  createRoute({
    getParentRoute: () => AdminRoute,
    path: '/orders/$id',
    component: AdminOrderDetail,
  }),
  // Products routes
  createRoute({
    getParentRoute: () => AdminRoute,
    path: '/products',
    component: AdminProducts,
  }),
  createRoute({
    getParentRoute: () => AdminRoute,
    path: '/products/$id',
    component: () => <div>Admin Product Detail - Coming Soon...</div>,
  }),
  // Customers routes
  createRoute({
    getParentRoute: () => AdminRoute,
    path: '/customers',
    component: AdminCustomers,
  }),
  createRoute({
    getParentRoute: () => AdminRoute,
    path: '/customers/$id',
    component: () => <div>Admin Customer Detail - Coming Soon...</div>,
  }),
  // Analytics routes
  createRoute({
    getParentRoute: () => AdminRoute,
    path: '/analytics',
    component: AdminAnalytics,
  }),
  // Marketing placeholder
  createRoute({
    getParentRoute: () => AdminRoute,
    path: '/marketing',
    component: AdminMarketing,
  }),
  // Content placeholder
  createRoute({
    getParentRoute: () => AdminRoute,
    path: '/content',
    component: AdminContent,
  }),
  // Settings placeholder
  createRoute({
    getParentRoute: () => AdminRoute,
    path: '/settings',
    component: AdminSettings,
  }),
]);
**/


const AdminRouteWithChildren = AdminRoute.addChildren([
  AdminDashboardRoute,
  AdminOrdersRoute,
  AdminOrderDetailRoute,
  AdminProductsRoute,
  AdminProductDetailRoute,
  AdminCustomersRoute,
  AdminCustomerDetailRoute,
  AdminAnalyticsRoute,
  AdminMarketingRoute,
  AdminContentRoute,
  AdminSettingsRoute,
]);


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
    AccountRoute.addChildren([
      AccountDashboardRoute,
      AccountOrdersRoute.addChildren([AccountOrderDetailRoute]),
      AccountAddressesRoute,
      AccountWishlistRoute,
      AccountSettingsRoute,
    ]),
    AdminRouteWithChildren,
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
