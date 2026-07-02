import { useEffect, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
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

  // ── verifying / pending: "One moment." ──
  if (status === 'verifying' || status === 'pending') {
    return (
      <div className="checkout-status-card">
        <h1>One moment.</h1>
        <p>
          Confirming your payment with Paystack…
          {status === 'pending' && pollCount > 0 && (
            <> ({pollCount}/15)</>
          )}
        </p>
      </div>
    );
  }

  // ── timeout: still processing, ambiguous ──
  if (status === 'timeout') {
    return (
      <div className="checkout-status-card">
        <h1>One moment.</h1>
        <p>
          Your payment is still being processed. Check your email for confirmation — we'll send your order details shortly.
        </p>
        <button className="btn btn-primary" onClick={() => navigate({ to: '/account' })}>
          My account
        </button>
      </div>
    );
  }

  // ── failed ──
  if (status === 'failed') {
    return (
      <div className="checkout-status-card">
        <h1>Payment didn't complete.</h1>
        <p>
          {!reference
            ? 'No payment reference found. Please contact support if you believe this is an error.'
            : "We couldn't verify your payment. Please contact support if you believe this is an error."}
        </p>
        <button className="btn btn-primary" onClick={() => navigate({ to: '/checkout' })}>
          Back to checkout
        </button>
      </div>
    );
  }

  // ── success ──
  return (
    <div className="checkout-status-card">
      <h1>Thank you.</h1>
      <p>Order {reference} is confirmed. A receipt is on its way to your inbox.</p>
      <button className="btn btn-primary" onClick={() => navigate({ to: '/account/orders' })}>
        View your orders
      </button>
    </div>
  );
}
