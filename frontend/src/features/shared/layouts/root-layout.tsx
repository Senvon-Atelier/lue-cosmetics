import { useState } from 'react';
import { Outlet } from '@tanstack/react-router';
import { AnnouncementBar } from './announcement-bar';
import { Header } from './header';
import { Footer } from './footer';
import { CartDrawer } from '../../cart/cart-drawer';

export function RootLayout() {
  const [isCartOpen, setIsCartOpen] = useState(false);

  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      <AnnouncementBar />
      <Header onCartOpen={() => setIsCartOpen(true)} />
      <main>
        <Outlet />
      </main>
      <Footer />
      <CartDrawer open={isCartOpen} onClose={() => setIsCartOpen(false)} />
    </div>
  );
}
