import { useState, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { useCart } from '../cart/cart-provider';
import { formatGhs } from '../../lib/format/utils';
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
  const { items, itemCount, subtotalGhsMinor } = useCart();
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
      navigate({ to: '/login', search: { redirect: '/checkout' } });
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

      if (response?.authorization_url) {
        // Redirect to Paystack
        window.location.href = response.authorization_url;
      } else {
        setErrors({ general: 'Unable to start checkout. Please try again.' });
      }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
      <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <span style={{ color: 'var(--ink-muted)', fontFamily: 'var(--font-label)' }}>Loading…</span>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  const shippingCost = 2500;
  const total = (subtotalGhsMinor || 0) + shippingCost;

  return (
    <main className="checkout-page fade-up">
      <div className="eyebrow">Checkout</div>
      <h1 className="h-display" style={{ fontSize: 'clamp(32px, 4vw, 56px)', marginBottom: '8px' }}>
        Almost <em>there.</em>
      </h1>

      <form onSubmit={handleSubmit}>
        {errors.general && (
          <div className="checkout-error-banner">{errors.general}</div>
        )}

        {/* ── Section 1: Delivery address ── */}
        <div className="checkout-section">
          <h2>Delivery address</h2>

          <div className="field">
            <label>Street Address <span aria-hidden="true">*</span></label>
            <input
              type="text"
              value={formData.line1}
              onChange={(e) => handleInputChange('line1', e.target.value)}
              placeholder="House number, street name"
              disabled={isProcessing}
            />
            {errors.line1 && <span className="field-error">{errors.line1}</span>}
          </div>

          <div className="field" style={{ marginTop: '16px' }}>
            <label>Apartment, suite, etc. (optional)</label>
            <input
              type="text"
              value={formData.line2}
              onChange={(e) => handleInputChange('line2', e.target.value)}
              placeholder="Apartment, suite, building (optional)"
              disabled={isProcessing}
            />
          </div>

          <div className="field" style={{ marginTop: '16px' }}>
            <label>City <span aria-hidden="true">*</span></label>
            <input
              type="text"
              value={formData.city}
              onChange={(e) => handleInputChange('city', e.target.value)}
              placeholder="Accra"
              disabled={isProcessing}
            />
            {errors.city && <span className="field-error">{errors.city}</span>}
          </div>

          <div className="field" style={{ marginTop: '16px' }}>
            <label>Region <span aria-hidden="true">*</span></label>
            <input
              type="text"
              value={formData.region}
              onChange={(e) => handleInputChange('region', e.target.value)}
              placeholder="Greater Accra"
              disabled={isProcessing}
            />
            {errors.region && <span className="field-error">{errors.region}</span>}
          </div>

          <div className="field" style={{ marginTop: '16px' }}>
            <label>Phone Number <span aria-hidden="true">*</span></label>
            <input
              type="tel"
              value={formData.phone}
              onChange={(e) => handleInputChange('phone', e.target.value)}
              placeholder="0XX XXX XXXX"
              disabled={isProcessing}
            />
            {errors.phone && <span className="field-error">{errors.phone}</span>}
          </div>

          <div className="field" style={{ marginTop: '16px' }}>
            <label>Address Label (optional)</label>
            <input
              type="text"
              value={formData.label}
              onChange={(e) => handleInputChange('label', e.target.value)}
              placeholder="Home, Work, etc."
              disabled={isProcessing}
            />
          </div>
        </div>

        {/* ── Section 2: Delivery method ── */}
        <div className="checkout-section">
          <h2>Delivery method</h2>
          <div className="checkout-methods">
            <label className="method-card selected">
              <input
                type="radio"
                name="shipping_method"
                value="standard"
                defaultChecked
                style={{ position: 'absolute', opacity: 0 }}
                readOnly
              />
              <span>Standard Delivery</span>
              <span className="price">{formatGhs(shippingCost)}</span>
            </label>
          </div>
        </div>

        {/* ── Section 3: Order summary ── */}
        <div className="checkout-section">
          <h2>Order summary</h2>
          <div className="checkout-summary-rows">
            {items.map((item) => (
              <div key={item.id} className="drawer-row">
                <span>{item.product_name} × {item.qty}</span>
                <span>{formatGhs(item.line_total_ghs_minor ?? 0)}</span>
              </div>
            ))}
            <div className="drawer-row muted">
              <span>Subtotal</span>
              <span>{formatGhs(subtotalGhsMinor)}</span>
            </div>
            <div className="drawer-row muted">
              <span>Delivery</span>
              <span>{formatGhs(shippingCost)}</span>
            </div>
            <div className="drawer-row">
              <span>Total</span>
              <span className="price">{formatGhs(total)}</span>
            </div>
          </div>
        </div>

        <button
          className="btn btn-primary"
          type="submit"
          disabled={isProcessing}
          style={{ width: '100%', justifyContent: 'center' }}
        >
          {isProcessing ? 'Preparing secure payment…' : `Pay with Paystack · ${formatGhs(total)}`}
        </button>
      </form>
    </main>
  );
}
