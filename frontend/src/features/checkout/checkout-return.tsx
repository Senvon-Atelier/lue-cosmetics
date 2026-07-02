import { useEffect, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';
import { Button } from '../shared/ui/button';
import { getCheckoutVerifyReference } from '../../lib/api/generated/rueCosmeticsAPI';
import { useCart } from '../cart/cart-provider';

export function CheckoutReturnPage() {
  const navigate = useNavigate();
  const [searchParams] = useState(() => new URLSearchParams(window.location.search));
  const reference = searchParams.get('reference');
  const [status, setStatus] = useState<'verifying' | 'success' | 'failed' | 'pending' | 'timeout'>('verifying');
  const [pollCount, setPollCount] = useState(0);
  const { refreshCart } = useCart();

  useEffect(() => {
    if (!reference) {
      setStatus('failed');
      return;
    }

    let isMounted = true;
    const maxPolls = 15; // 15 polls * 2s = 30s total polling budget
    let currentPoll = 0;

    const verifyPayment = async () => {
      try {
        const response = await getCheckoutVerifyReference(reference);

        if (!isMounted) return;

        if (response?.status === 'paid') {
          setStatus('success');
          await refreshCart();
          return;
        }

        if (response?.status === 'pending') {
          currentPoll++;
          setPollCount(currentPoll);

          if (currentPoll < maxPolls) {
            // Poll again in 2 seconds
            setTimeout(verifyPayment, 2000);
          } else {
            // Exceeded polling budget
            setStatus('timeout');
          }
        } else {
          // Unexpected status
          setStatus('failed');
        }
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } catch (error: any) {
        console.error('Payment verification failed:', error);
        if (isMounted) {
          setStatus('failed');
        }
      }
    };

    verifyPayment();

    return () => {
      isMounted = false;
    };
  }, [reference]);

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

  if (status === 'pending') {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-center">
          <div className="w-16 h-16 border-4 border-lavender-600 border-t-transparent rounded-full animate-spin mb-4 mx-auto" />
          <h2 className="font-display text-2xl mb-2">Processing payment...</h2>
          <p className="text-ink-muted">
            Still verifying... {pollCount}/15
          </p>
        </div>
      </div>
    );
  }

  if (status === 'timeout') {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-center max-w-md">
          <div className="w-16 h-16 rounded-full bg-yellow-100 flex items-center justify-center mx-auto mb-4">
            <Icon name="clock" size={32} className="text-yellow-600" />
          </div>
          <h2 className="font-display text-2xl mb-2">Still Processing</h2>
          <p className="text-ink-muted mb-6">
            Your payment is still being processed. Check your email for confirmation. We'll send your order details shortly.
          </p>
          <div className="flex gap-3 justify-center">
            <Button onClick={() => navigate({ to: '/account' })}>My Account</Button>
            <Button variant="outline" onClick={() => navigate({ to: '/shop' })}>
              Continue Shopping
            </Button>
          </div>
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
            {!reference
              ? 'No payment reference found. Please contact support if you believe this is an error.'
              : 'We couldn\'t verify your payment. Please contact support if you believe this is an error.'}
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

  // Get order total from session storage (stored by checkout page)
  const orderTotal = sessionStorage.getItem('order_total');

  return (
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
            <span className="font-label font-semibold">{orderTotal} GHS</span>
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
  );
}
