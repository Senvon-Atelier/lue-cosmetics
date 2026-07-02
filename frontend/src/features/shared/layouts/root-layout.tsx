import { Outlet } from '@tanstack/react-router';
import { AnnouncementBar } from './announcement-bar';
import { Header } from './header';
import { Footer } from './footer';
import { CartDrawer } from '../../cart/cart-drawer';
import { AddToast } from '../../cart/add-toast';
import { useCart } from '../../cart/cart-provider';

export function RootLayout() {
  const { isDrawerOpen, openDrawer, closeDrawer, lastAdded, dismissToast } = useCart();

  return (
    <div>
      <AnnouncementBar />
      <Header />
      <main>
        <Outlet />
      </main>
      <Footer />
      <CartDrawer open={isDrawerOpen} onClose={closeDrawer} />
      <AddToast lastAdded={lastAdded} onView={openDrawer} onDismiss={dismissToast} />
    </div>
  );
}
