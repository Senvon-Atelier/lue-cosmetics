import type { ReactNode } from 'react';
import { createRouter, createRoute, createRootRoute, Outlet, redirect, useParams } from '@tanstack/react-router';
import { QueryProvider, queryClient } from './features/shared/providers/query-provider';
import { AuthProvider } from './lib/auth/auth-provider';
import { CartProvider } from './features/cart/cart-provider';
import { sessionQueryOptions } from './lib/auth/session';
import { redirectPathFor, GuardRequirement } from './lib/auth/guards';
import { RootLayout, CheckoutLayout } from './features/shared/layouts';
import { HomePage } from './features/home/home-page';
import { ShopPage } from './features/catalog/shop-page';
import { ProductDetail } from './features/catalog/product-detail';
import { CartPage } from './features/cart/cart-page';
import { CheckoutPage } from './features/checkout/checkout-page';
import { CheckoutReturnPage } from './features/checkout/checkout-return';
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

// Session-backed route guard. Throws a redirect before anything renders.
async function requireRole(requirement: GuardRequirement) {
  const session = await queryClient.ensureQueryData(sessionQueryOptions);
  const target = redirectPathFor(session, requirement);
  if (target) throw redirect({ to: target });
}

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

// ── Storefront: public pages + auth pages (Header + Footer + CartDrawer) ────
const storefrontLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: '_storefront',
  component: RootLayout,
});

const pageShell = (children: ReactNode) => (
  <div>{children}</div>
);

const homeRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/',
  component: HomePage,
});

const shopRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/shop',
  component: () => pageShell(<ShopPage />),
});

function ProductDetailComponent() {
  const { slug } = useParams({ from: '/_storefront/shop/$slug' });
  return pageShell(<ProductDetail slug={slug || ''} />);
}

const productDetailRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/shop/$slug',
  component: ProductDetailComponent,
});

const cartRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/cart',
  component: () => pageShell(<CartPage />),
});

const aboutRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/about',
  component: () => pageShell(<AboutPage />),
});

// ── Auth: no storefront chrome ───────────────────────────────────────────────
const authLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: '_auth',
  component: () => <Outlet />,
});

const loginRoute = createRoute({ getParentRoute: () => authLayoutRoute, path: '/login', component: LoginPage });
const signupRoute = createRoute({ getParentRoute: () => authLayoutRoute, path: '/signup', component: SignupPage });
const forgotPasswordRoute = createRoute({ getParentRoute: () => authLayoutRoute, path: '/forgot-password', component: ForgotPasswordPage });
const resetPasswordRoute = createRoute({ getParentRoute: () => authLayoutRoute, path: '/reset-password', component: ResetPasswordPage });
const verifyEmailRoute = createRoute({ getParentRoute: () => authLayoutRoute, path: '/verify-email', component: VerifyEmailPage });

// ── Account: authenticated customers, storefront chrome ─────────────────────
const accountRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/account',
  beforeLoad: () => requireRole('authenticated'),
  component: AccountLayout,
});

const accountDashboardRoute = createRoute({ getParentRoute: () => accountRoute, path: '/', component: AccountDashboard });
const accountOrdersRoute = createRoute({ getParentRoute: () => accountRoute, path: '/orders', component: AccountOrders });
const accountOrderDetailRoute = createRoute({ getParentRoute: () => accountOrdersRoute, path: '$id', component: AccountOrderDetail });
const accountAddressesRoute = createRoute({ getParentRoute: () => accountRoute, path: '/addresses', component: AccountAddresses });
const accountWishlistRoute = createRoute({ getParentRoute: () => accountRoute, path: '/wishlist', component: AccountWishlist });
const accountSettingsRoute = createRoute({ getParentRoute: () => accountRoute, path: '/settings', component: AccountSettings });

// ── Admin: admin role required, own chrome (no storefront header/footer) ────
const adminRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/admin',
  beforeLoad: () => requireRole('admin'),
  component: AdminLayout,
});

const adminDashboardRoute = createRoute({ getParentRoute: () => adminRoute, path: '/', component: AdminDashboard });
const adminOrdersRoute = createRoute({ getParentRoute: () => adminRoute, path: '/orders', component: AdminOrders });
const adminOrderDetailRoute = createRoute({ getParentRoute: () => adminRoute, path: '/orders/$id', component: AdminOrderDetail });
const adminProductsRoute = createRoute({ getParentRoute: () => adminRoute, path: '/products', component: AdminProducts });
const adminProductDetailRoute = createRoute({ getParentRoute: () => adminRoute, path: '/products/$id', component: () => <div>Admin Product Detail - Coming Soon...</div> });
const adminCustomersRoute = createRoute({ getParentRoute: () => adminRoute, path: '/customers', component: AdminCustomers });
const adminCustomerDetailRoute = createRoute({ getParentRoute: () => adminRoute, path: '/customers/$id', component: () => <div>Admin Customer Detail - Coming Soon...</div> });
const adminAnalyticsRoute = createRoute({ getParentRoute: () => adminRoute, path: '/analytics', component: AdminAnalytics });
const adminMarketingRoute = createRoute({ getParentRoute: () => adminRoute, path: '/marketing', component: AdminMarketing });
const adminContentRoute = createRoute({ getParentRoute: () => adminRoute, path: '/content', component: AdminContent });
const adminSettingsRoute = createRoute({ getParentRoute: () => adminRoute, path: '/settings', component: AdminSettings });

// ── Checkout: minimal chrome ─────────────────────────────────────────────────
const checkoutLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: '_checkout',
  component: CheckoutLayout,
});

const checkoutRoute = createRoute({ getParentRoute: () => checkoutLayoutRoute, path: '/checkout', component: CheckoutPage });
const checkoutReturnRoute = createRoute({
  getParentRoute: () => checkoutLayoutRoute,
  path: '/checkout/return',
  component: () => pageShell(<CheckoutReturnPage />),
});

const routeTree = rootRoute.addChildren([
  storefrontLayoutRoute.addChildren([
    homeRoute,
    shopRoute,
    productDetailRoute,
    cartRoute,
    aboutRoute,
    accountRoute.addChildren([
      accountDashboardRoute,
      accountOrdersRoute.addChildren([accountOrderDetailRoute]),
      accountAddressesRoute,
      accountWishlistRoute,
      accountSettingsRoute,
    ]),
  ]),
  authLayoutRoute.addChildren([
    loginRoute,
    signupRoute,
    forgotPasswordRoute,
    resetPasswordRoute,
    verifyEmailRoute,
  ]),
  adminRoute.addChildren([
    adminDashboardRoute,
    adminOrdersRoute,
    adminOrderDetailRoute,
    adminProductsRoute,
    adminProductDetailRoute,
    adminCustomersRoute,
    adminCustomerDetailRoute,
    adminAnalyticsRoute,
    adminMarketingRoute,
    adminContentRoute,
    adminSettingsRoute,
  ]),
  checkoutLayoutRoute.addChildren([checkoutRoute, checkoutReturnRoute]),
]);

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
  defaultPreloadStaleTime: 0,
});

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
