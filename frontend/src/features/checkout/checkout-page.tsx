import { useState, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { useCart } from '../cart/cart-provider';
import { Button } from '../shared/ui/button';
import { formatPrice } from '../../lib/format/utils';
import { postCheckoutInit } from '../../lib/api/generated/rueCosmeticsAPI';
import type { InternalOrdersInitCheckoutBody } from '../../lib/api/generated/rueCosmeticsAPI';

interface FormErrors {
  line1?: string;
  city?: string;
  region?: string;
  phone?: string;
  general?: string;
}

export function CheckoutPage() {
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();
  const { itemCount, subtotalGhsMinor } = useCart();
  const navigate = useNavigate();
  const [isProcessing, setIsProcessing] = useState(false);
  const [errors, setErrors] = useState<FormErrors>({});

  // Form state
  const [formData, setFormData] = useState({
    line1: '',
    line2: '',
    city: '',
    region: '',
    phone: '',
    label: 'Home',
  });

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

  const validateForm = (): boolean => {
    const newErrors: FormErrors = {};

    if (!formData.line1.trim()) {
      newErrors.line1 = 'Street address is required';
    }

    if (!formData.city.trim()) {
      newErrors.city = 'City is required';
    }

    if (!formData.region.trim()) {
      newErrors.region = 'Region is required';
    }

    if (!formData.phone.trim()) {
      newErrors.phone = 'Phone number is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    if (!user?.user_id) {
      setErrors({ general: 'Unable to process checkout. Please try logging in again.' });
      return;
    }

    setIsProcessing(true);
    setErrors({});

    try {
      const checkoutBody: InternalOrdersInitCheckoutBody = {
        shipping_address: {
          line1: formData.line1.trim(),
          line2: formData.line2.trim() || undefined,
          city: formData.city.trim(),
          region: formData.region.trim(),
          phone: formData.phone.trim(),
          label: formData.label.trim() || undefined,
        },
        shipping_method: 'standard',
      };

      const response = await postCheckoutInit(checkoutBody);

      if (response.data?.authorization_url) {
        // Store total in session storage for the return page
        sessionStorage.setItem('order_total', String(subtotalGhsMinor || 0));
        // Redirect to Paystack
        window.location.href = response.data.authorization_url;
      } else {
        setErrors({ general: 'Unable to start checkout. Please try again.' });
      }
    } catch (error: any) {
      console.error('Checkout failed:', error);

      if (error.response?.status === 400) {
        const errorData = error.response.data;

        if (errorData?.error?.code === 'validation') {
          const detail = errorData?.error?.detail || '';
          if (detail.includes('empty cart') || detail.includes('cart is empty')) {
            setErrors({ general: 'Your cart is empty. Please add items before checkout.' });
          } else {
            setErrors({ general: 'Please check your shipping address. Some fields may be invalid.' });
          }
        } else {
          setErrors({ general: 'Please check your shipping address.' });
        }
      } else if (error.response?.status === 503) {
        setErrors({
          general: 'Payments are temporarily unavailable. Please try again in a few minutes.',
        });
      } else {
        setErrors({ general: 'Unable to start checkout. Please try again.' });
      }
    } finally {
      setIsProcessing(false);
    }
  };

  const handleInputChange = (field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    // Clear error for this field when user starts typing
    if (errors[field as keyof FormErrors]) {
      setErrors((prev) => ({ ...prev, [field]: undefined }));
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-ink-muted">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
          <h1 className="font-display text-4xl mb-4">Login Required</h1>
          <p className="text-ink-muted mb-6">Please log in to proceed with checkout.</p>
          <Button onClick={() => navigate({ to: '/login' })}>Go to Login</Button>
        </div>
      </div>
    );
  }

  const shippingCost = 2500;
  const total = (subtotalGhsMinor || 0) + shippingCost;

  return (
    <>
      <div className="mb-6">
        <h1 className="font-display text-4xl mb-2">Checkout</h1>
        <p className="text-ink-muted">Complete your delivery details below</p>
      </div>

      <div className="grid md:grid-cols-3 gap-8">
        {/* Main Content */}
        <div className="md:col-span-2 space-y-6">
          {/* Delivery Address Form */}
          <div className="bg-lavender-50 rounded-lg p-6">
            <h2 className="font-display text-2xl mb-4">Delivery Address</h2>

            <form onSubmit={handleSubmit} className="space-y-4">
              {/* General Error */}
              {errors.general && (
                <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-4">
                  {errors.general}
                </div>
              )}

              {/* Line 1 (Required) */}
              <div>
                <label className="block font-label font-medium text-sm mb-2">
                  Street Address <span className="text-red-600">*</span>
                </label>
                <input
                  type="text"
                  value={formData.line1}
                  onChange={(e) => handleInputChange('line1', e.target.value)}
                  className={`w-full px-4 py-3 border rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-lavender-600 focus:border-transparent ${
                    errors.line1 ? 'border-red-500' : 'border-line-soft'
                  }`}
                  placeholder="House number, street name"
                  disabled={isProcessing}
                />
                {errors.line1 && <p className="text-red-600 text-sm mt-1">{errors.line1}</p>}
              </div>

              {/* Line 2 (Optional) */}
              <div>
                <label className="block font-label font-medium text-sm mb-2">Apartment, suite, etc. (optional)</label>
                <input
                  type="text"
                  value={formData.line2}
                  onChange={(e) => handleInputChange('line2', e.target.value)}
                  className="w-full px-4 py-3 border border-line-soft rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-lavender-600 focus:border-transparent"
                  placeholder="Apartment, suite, building (optional)"
                  disabled={isProcessing}
                />
              </div>

              {/* City (Required) */}
              <div>
                <label className="block font-label font-medium text-sm mb-2">
                  City <span className="text-red-600">*</span>
                </label>
                <input
                  type="text"
                  value={formData.city}
                  onChange={(e) => handleInputChange('city', e.target.value)}
                  className={`w-full px-4 py-3 border rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-lavender-600 focus:border-transparent ${
                    errors.city ? 'border-red-500' : 'border-line-soft'
                  }`}
                  placeholder="Accra"
                  disabled={isProcessing}
                />
                {errors.city && <p className="text-red-600 text-sm mt-1">{errors.city}</p>}
              </div>

              {/* Region (Required) */}
              <div>
                <label className="block font-label font-medium text-sm mb-2">
                  Region <span className="text-red-600">*</span>
                </label>
                <input
                  type="text"
                  value={formData.region}
                  onChange={(e) => handleInputChange('region', e.target.value)}
                  className={`w-full px-4 py-3 border rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-lavender-600 focus:border-transparent ${
                    errors.region ? 'border-red-500' : 'border-line-soft'
                  }`}
                  placeholder="Greater Accra"
                  disabled={isProcessing}
                />
                {errors.region && <p className="text-red-600 text-sm mt-1">{errors.region}</p>}
              </div>

              {/* Phone (Required) */}
              <div>
                <label className="block font-label font-medium text-sm mb-2">
                  Phone Number <span className="text-red-600">*</span>
                </label>
                <input
                  type="tel"
                  value={formData.phone}
                  onChange={(e) => handleInputChange('phone', e.target.value)}
                  className={`w-full px-4 py-3 border rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-lavender-600 focus:border-transparent ${
                    errors.phone ? 'border-red-500' : 'border-line-soft'
                  }`}
                  placeholder="0XX XXX XXXX"
                  disabled={isProcessing}
                />
                {errors.phone && <p className="text-red-600 text-sm mt-1">{errors.phone}</p>}
              </div>

              {/* Label (Optional) */}
              <div>
                <label className="block font-label font-medium text-sm mb-2">Address Label (optional)</label>
                <input
                  type="text"
                  value={formData.label}
                  onChange={(e) => handleInputChange('label', e.target.value)}
                  className="w-full px-4 py-3 border border-line-soft rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-lavender-600 focus:border-transparent"
                  placeholder="Home, Work, etc."
                  disabled={isProcessing}
                />
              </div>

              {/* Submit Button */}
              <Button
                type="submit"
                disabled={isProcessing}
                isLoading={isProcessing}
                className="w-full"
                icon="arrow"
                iconPosition="right"
              >
                Proceed to Payment
              </Button>
            </form>
          </div>

          {/* Shipping Method */}
          <div className="bg-lavender-50 rounded-lg p-6">
            <h2 className="font-display text-2xl mb-4">Shipping Method</h2>

            <div className="space-y-3">
              <label
                className={`flex items-center justify-between p-4 border rounded-lg cursor-pointer transition-colors bg-lavender-100 border-lavender-600`}
              >
                <div className="flex items-center gap-3">
                  <input
                    type="radio"
                    name="shipping"
                    value="standard"
                    checked
                    readOnly
                    className="w-5 h-5 text-lavender-600"
                    disabled
                  />
                  <div>
                    <div className="font-label font-semibold">Standard Delivery</div>
                    <div className="text-sm text-ink-muted">3-5 business days · GHS 25.00</div>
                  </div>
                </div>
                <div className="text-sm font-semibold">{formatPrice(shippingCost)}</div>
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
                <span className="font-label font-semibold">{formatPrice(subtotalGhsMinor)}</span>
              </div>
              <div className="flex justify-between text-sm text-ink-muted">
                <span>Delivery</span>
                <span>{formatPrice(shippingCost)}</span>
              </div>
              <div className="border-t border-line-soft pt-3 mt-3">
                <div className="flex justify-between">
                  <span className="font-label font-semibold">Total</span>
                  <span className="font-display text-xl">{formatPrice(total)}</span>
                </div>
              </div>
            </div>

            <div className="mt-4 text-xs text-ink-muted text-center">
              <p>Secure checkout powered by Paystack</p>
              <p className="mt-1">Delivery calculated at checkout</p>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
