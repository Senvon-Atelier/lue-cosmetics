import { Outlet } from '@tanstack/react-router';
import { Brand } from '../ui/brand';

export function CheckoutLayout() {
  return (
    <div>
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
        <Brand />
      </div>
      <div>
        <Outlet />
      </div>
    </div>
  );
}
