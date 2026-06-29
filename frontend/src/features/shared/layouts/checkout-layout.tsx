import { Outlet } from '@tanstack/react-router';
import { Brand } from '../ui/brand';

export function CheckoutLayout() {
  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
        <Brand />
      </div>
      <main>
        <Outlet />
      </main>
    </div>
  );
}
