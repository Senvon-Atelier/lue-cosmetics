import ReactDOM from 'react-dom/client';
import { StrictMode } from 'react';
import './styles/globals.css';
import { QueryProvider } from './features/shared/providers/query-provider';
import { AuthProvider } from './lib/auth/auth-provider';
import { CartProvider } from './features/cart/cart-provider';
import { Brand, Button } from './features/shared/ui';

// Test component to verify providers are working
function App() {
  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
        <header className="mb-8">
          <Brand />
        </header>

        <main>
          <h1 className="font-display text-4xl mb-4">Rue Cosmetics</h1>
          <p className="text-ink-muted mb-6">
            Frontend foundation is complete. Phase 3 (API Client & Auth) is wired up.
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
              </ul>
            </div>

            <div className="p-4 bg-lavender-100 rounded">
              <h2 className="font-label font-semibold mb-2">🚧 Next Steps</h2>
              <ul className="space-y-1 text-ink-soft">
                <li>• Phase 4: TanStack Router setup with protected routes</li>
                <li>• Phase 5: Product catalog features (ProductCard, ShopPage)</li>
                <li>• Phase 6: Cart & checkout flow</li>
                <li>• Phase 7: Home page & marketing content</li>
                <li>• Phase 8: Testing & deployment</li>
              </ul>
            </div>

            <div className="flex gap-2">
              <Button variant="primary">Shop Now</Button>
              <Button variant="outline">Learn More</Button>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryProvider>
      <AuthProvider>
        <CartProvider>
          <App />
        </CartProvider>
      </AuthProvider>
    </QueryProvider>
  </StrictMode>,
);
