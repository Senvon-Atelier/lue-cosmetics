import { useState, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { useCart } from '../cart/cart-provider';
import { Button } from '../shared/ui/button';
import { formatPrice } from '../../lib/format/utils';
import { postCheckoutInit } from '../../lib/api/checkout-api';

export function CheckoutPage() {
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();
  const { itemCount, subtotalGhsMinor } = useCart();
  const navigate = useNavigate();
  const [selectedAddress, setSelectedAddress] = useState<string | null>(null);
  const [shippingMethod, setShippingMethod] = useState('standard');
  const [isProcessing, setIsProcessing] = useState(false);

  // Demo addresses
  const addresses = [
    { id: 'addr-1', label: 'Home', line1: 'Community 18, Spintex Road', line2: 'Adjacent KFC', city: 'Accra', region: 'Greater Accra', phone: '0594 701 345', isDefault: true },
    { id: 'addr-2', label: 'Work', line1: 'Airport Industrial Area', line2: 'Plot 123', city: 'Accra', region: 'Greater Accra', phone: '020 123 4567', isDefault: false },
  ];

  // Redirect if not authenticated
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      navigate({ to: '/login' });
    }
  }, [authLoading, isAuthenticated, navigate]);

  // Redirect if cart is empty
  useEffect(() => {
    if (!authLoading && isAuthenticated && itemCount === 0) {
      navigate({ to: '/shop' });
    }
  }, [authLoading, isAuthenticated, itemCount, navigate]);

  const handleCheckout = async () => {
    if (!selectedAddress) {
      alert('Please select a delivery address');
      return;
    }

    if (!user?.user_id) {
      alert('Unable to process checkout. Please try logging in again.');
      return;
    }

    setIsProcessing(true);
    try {
      const response = await postCheckoutInit({
        address_id: selectedAddress,
        shipping_method: shippingMethod,
      });

      // Redirect to Paystack
      if (response.data && typeof response.data === 'object' && 'authorization_url' in response.data) {
        const { authorization_url } = response.data as { authorization_url?: string };
        if (authorization_url) {
          window.location.href = authorization_url;
        }
      }
    } catch (error) {
      console.error('Checkout failed:', error);
      alert('Unable to initiate checkout. Please try again.');
    } finally {
      setIsProcessing(false);
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-ink-muted">Loading...</div>
      </div>
    );
  }

  return (
    <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
      <div className="mb-6">
        <h1 className="font-display text-4xl mb-2">Checkout</h1>
        <p className="text-ink-muted">Complete your order details below</p>
      </div>

      <div className="grid md:grid-cols-3 gap-8">
        {/* Main Content */}
        <div className="md:col-span-2 space-y-6">
          {/* Delivery Address */}
          <div className="bg-lavender-50 rounded-lg p-6">
            <h2 className="font-display text-2xl mb-4">Delivery Address</h2>

            <div className="space-y-3">
              {addresses.map((address) => (
                <button
                  key={address.id}
                  onClick={() => setSelectedAddress(address.id)}
                  className={`w-full text-left p-4 border rounded-lg transition-colors ${
                    selectedAddress === address.id
                      ? 'border-lavender-600 bg-lavender-100'
                      : 'border-line-soft hover:border-lavender-300'
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-label font-semibold">{address.label}</span>
                    {address.isDefault && (
                      <span className="px-2 py-1 bg-lavender-200 text-xs rounded">Default</span>
                    )}
                  </div>
                  <div className="text-sm text-ink-soft">
                    <p>{address.line1}</p>
                    {address.line2 && <p>{address.line2}</p>}
                    <p>{address.city}, {address.region}</p>
                    <p>{address.phone}</p>
                  </div>
                </button>
              ))}
            </div>

            <button className="mt-4 text-sm text-lavender-600 hover:text-lavender-700 font-label underline">
              + Add new address
            </button>
          </div>

          {/* Shipping Method */}
          <div className="bg-lavender-50 rounded-lg p-6">
            <h2 className="font-display text-2xl mb-4">Shipping Method</h2>

            <div className="space-y-3">
              <label
                className={`flex items-center justify-between p-4 border rounded-lg cursor-pointer transition-colors ${
                  shippingMethod === 'standard'
                    ? 'border-lavender-600 bg-lavender-100'
                    : 'border-line-soft hover:border-lavender-300'
                }`}
              >
                <div className="flex items-center gap-3">
                  <input
                    type="radio"
                    name="shipping"
                    value="standard"
                    checked={shippingMethod === 'standard'}
                    onChange={(e) => setShippingMethod(e.target.value)}
                    className="w-5 h-5 text-lavender-600"
                  />
                  <div>
                    <div className="font-label font-semibold">Standard Delivery</div>
                    <div className="text-sm text-ink-muted">3-5 business days · GHS 25.00</div>
                  </div>
                </div>
                <div className="text-sm font-semibold">{formatPrice(2500)}</div>
              </label>
            </div>
          </div>
        </div>

        {/* Order Summary */}
        <div className="md:col-span-1">
          <div className="bg-lavender-50 rounded-lg p-6 sticky top-4">
            <h2 className="font-display text-2xl mb-4">Order Summary</h2>

            <div className="space-y-3 mb-6">
              <div className="flex justify-between text-sm">
                <span className="text-ink-muted">Subtotal</span>
                <span className="font-label">{formatPrice(subtotalGhsMinor)}</span>
              </div>
              <div className="flex justify-between text-sm text-ink-muted">
                <span>Delivery</span>
                <span>{formatPrice(2500)}</span>
              </div>
              <div className="border-t border-line-soft pt-3 mt-3">
                <div className="flex justify-between">
                  <span className="font-label font-semibold">Total</span>
                  <span className="font-display text-xl">{formatPrice((subtotalGhsMinor || 0) + 2500)}</span>
                </div>
              </div>
            </div>

            <Button
              onClick={handleCheckout}
              disabled={!selectedAddress || isProcessing}
              isLoading={isProcessing}
              className="w-full"
              icon="arrow"
              iconPosition="right"
            >
              Proceed to Payment
            </Button>

            <div className="mt-4 text-xs text-ink-muted text-center">
              <p>Secure checkout powered by Paystack</p>
              <p className="mt-1">Delivery calculated at checkout</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
