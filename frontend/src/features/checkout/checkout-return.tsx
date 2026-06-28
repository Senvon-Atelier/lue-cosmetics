import { useEffect, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';
import { Button } from '../shared/ui/button';
import { getCheckoutVerify } from '../../lib/api/checkout-api';
import { useCart } from '../cart/cart-provider';
import { formatPrice } from '../../lib/format/utils';

export function CheckoutReturnPage() {
  const navigate = useNavigate();
  const [searchParams] = useState(() => new URLSearchParams(window.location.search));
  const reference = searchParams.get('reference') as string | null;
  const [status, setStatus] = useState<'verifying' | 'success' | 'failed'>('verifying');
  const [orderTotal, setOrderTotal] = useState<string>('');
  const { refreshCart } = useCart();

  useEffect(() => {
    if (!reference) {
      setStatus('failed');
      return;
    }

    const verifyPayment = async () => {
      try {
        const response = await getCheckoutVerify(reference);
        if (response.status === 200) {
          setStatus('success');
          const orderData = response.data as any;
          if (orderData?.total_ghs_minor) {
            setOrderTotal(formatPrice(orderData.total_ghs_minor));
          }
          // Refresh cart to clear items after successful payment
          await refreshCart();
        } else {
          setStatus('failed');
        }
      } catch (error) {
        console.error('Payment verification failed:', error);
        setStatus('failed');
      }
    };

    // Poll for verification
    verifyPayment();

    // If still verifying after 30 seconds, show failed state
    const timeout = setTimeout(() => {
      if (status === 'verifying') {
        setStatus('failed');
      }
    }, 30000);

    return () => clearTimeout(timeout);
  }, []);

  if (status === 'verifying') {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-center">
          <div className="w-16 h-16 border-4 border-lavender-600 border-t-transparent rounded-full animate-spin mb-4 mx-auto" />
          <h2 className="font-display text-2xl mb-2">Verifying your payment...</h2>
          <p className="text-ink-muted">Please wait while we confirm your order.</p>
        </div>
      </div>
    );
  }

  if (status === 'failed') {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-center max-w-md">
          <div className="w-16 h-16 rounded-full bg-red-100 flex items-center justify-center mx-auto mb-4">
            <Icon name="close" size={32} className="text-red-600" />
          </div>
          <h2 className="font-display text-2xl mb-2">Payment Verification Failed</h2>
          <p className="text-ink-muted mb-6">
            We couldn't verify your payment. Please contact support if you believe this is an error.
          </p>
          <div className="flex gap-3 justify-center">
            <Button onClick={() => navigate({ to: '/cart' })}>Back to Cart</Button>
            <Button variant="outline" onClick={() => navigate({ to: '/shop' })}>
              Continue Shopping
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
        <div className="max-w-lg mx-auto text-center pt-16">
          <div className="w-20 h-20 rounded-full bg-green-100 flex items-center justify-center mx-auto mb-6">
            <Icon name="check" size={40} className="text-green-600" />
          </div>

          <h1 className="font-display text-4xl mb-4">Order Confirmed!</h1>
          <p className="text-lg text-ink-soft mb-6">
            Thank you for your order. We've sent a confirmation email with your order details.
          </p>

          <div className="bg-lavender-50 rounded-lg p-6 mb-6 text-left">
            <div className="flex justify-between mb-2">
              <span className="text-ink-muted">Reference:</span>
              <span className="font-label font-mono">{reference}</span>
            </div>
            {orderTotal && (
              <div className="flex justify-between">
                <span className="text-ink-muted">Total Paid:</span>
                <span className="font-label font-semibold">{orderTotal}</span>
              </div>
            )}
          </div>

          <div className="space-y-3">
            <Button onClick={() => navigate({ to: '/account' })} className="w-full">
              View Order Details
            </Button>
            <Button onClick={() => navigate({ to: '/shop' })} variant="outline" className="w-full">
              Continue Shopping
            </Button>
          </div>

          <div className="mt-8 p-4 bg-cream rounded-lg">
            <h3 className="font-label font-semibold mb-2">What happens next?</h3>
            <ul className="text-sm text-ink-soft space-y-1">
              <li>• You'll receive an order confirmation email shortly</li>
              <li>• We'll prepare your items for delivery</li>
              <li>• You'll receive another email when your order ships</li>
              <li>• Track your order status in your account</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}
