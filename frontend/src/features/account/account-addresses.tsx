import { useState, useEffect } from 'react';
import {
  getMeAddresses,
  postMeAddresses,
  patchMeAddressesId,
  deleteMeAddressesId,
  postMeAddressesIdDefault,
} from '../../lib/api/generated/rueCosmeticsAPI';
import { Button } from '../shared/ui/button';

type Address = {
  id?: string;
  label?: string;
  line1?: string;
  line2?: string;
  city?: string;
  region?: string;
  phone?: string;
  is_default?: boolean;
  created_at?: string;
  updated_at?: string;
};

export function AccountAddresses() {
  const [addresses, setAddresses] = useState<Address[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingAddress, setEditingAddress] = useState<Address | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadAddresses = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await getMeAddresses();
      setAddresses(response.addresses || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load addresses');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadAddresses();
  }, []);

  const handleSetDefault = async (id: string) => {
    if (!id) return;
    try {
      await postMeAddressesIdDefault(id);
      await loadAddresses();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to set default address');
    }
  };

  const handleDelete = async (id: string) => {
    if (!id || !confirm('Are you sure you want to delete this address?')) return;

    try {
      await deleteMeAddressesId(id);
      await loadAddresses();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete address');
    }
  };

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <div className="text-ink-muted">Loading addresses...</div>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="font-display text-xl mb-2">Addresses</h2>
          <p className="text-ink-muted">Manage your shipping addresses.</p>
        </div>
        <Button variant="primary" onClick={() => setShowForm(true)}>
          Add New Address
        </Button>
      </div>

      {/* Error Message */}
      {error && (
        <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg mb-4">
          {error}
        </div>
      )}

      {/* Address Form */}
      {showForm && (
        <AddressForm
          address={editingAddress}
          onSubmit={async (data) => {
            try {
              if (editingAddress && editingAddress.id) {
                await patchMeAddressesId(editingAddress.id, data);
              } else {
                await postMeAddresses(data);
              }
              setShowForm(false);
              setEditingAddress(null);
              await loadAddresses();
            } catch (err) {
              setError(err instanceof Error ? err.message : 'Failed to save address');
            }
          }}
          onCancel={() => {
            setShowForm(false);
            setEditingAddress(null);
          }}
        />
      )}

      {/* Addresses List */}
      {addresses.length === 0 ? (
        <div className="text-center py-12">
          <div className="text-4xl mb-4">📍</div>
          <h3 className="font-display text-xl mb-2">No addresses yet</h3>
          <p className="text-ink-muted mb-6">Add a shipping address to make checkout easier.</p>
          <Button variant="primary" onClick={() => setShowForm(true)}>
            Add Your First Address
          </Button>
        </div>
      ) : (
        <div className="grid md:grid-cols-2 gap-4">
          {addresses.map((address) => (
            <div
              key={address.id}
              className="bg-white rounded-lg p-6"
              style={{ border: '1px solid var(--line)' }}
            >
              <div className="flex items-start justify-between mb-4">
                <div>
                  <h3 className="font-label font-semibold mb-1">{address.label || 'Address'}</h3>
                  <div className="text-ink-soft text-sm">
                    {address.line1}
                    {address.line2 && <div>{address.line2}</div>}
                    <div>{address.city && address.region ? `${address.city}, ${address.region}` : address.city || address.region}</div>
                    <div>{address.phone}</div>
                  </div>
                </div>
                {address.is_default && (
                  <span className="px-2 py-1 bg-lavender-100 text-lavender-700 text-xs font-label font-medium rounded">
                    Default
                  </span>
                )}
              </div>

              <div className="flex gap-2 mt-4">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => address.id && setEditingAddress(address)}
                >
                  Edit
                </Button>
                {!address.is_default && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => address.id && handleSetDefault(address.id)}
                  >
                    Set Default
                  </Button>
                )}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => address.id && handleDelete(address.id)}
                >
                  Delete
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// Address Form Component
function AddressForm({
  address,
  onSubmit,
  onCancel,
}: {
  address: Address | null;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  onSubmit: (data: any) => Promise<void>;
  onCancel: () => void;
}) {
  const [label, setLabel] = useState(address?.label || '');
  const [line1, setLine1] = useState(address?.line1 || '');
  const [line2, setLine2] = useState(address?.line2 || '');
  const [city, setCity] = useState(address?.city || '');
  const [region, setRegion] = useState(address?.region || '');
  const [phone, setPhone] = useState(address?.phone || '');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrors({});
    setIsSubmitting(true);

    // Validation
    const newErrors: Record<string, string> = {};
    if (!label.trim()) newErrors.label = 'Label is required';
    if (!line1.trim()) newErrors.line1 = 'Address line 1 is required';
    if (!city.trim()) newErrors.city = 'City is required';
    if (!region.trim()) newErrors.region = 'Region is required';
    if (!phone.trim()) newErrors.phone = 'Phone is required';

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      setIsSubmitting(false);
      return;
    }

    try {
      await onSubmit({
        label,
        line1,
        line2,
        city,
        region,
        phone,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="bg-white rounded-lg p-6 mb-6" style={{ border: '1px solid var(--line)' }}>
      <h3 className="font-label font-semibold mb-4">
        {address ? 'Edit Address' : 'Add New Address'}
      </h3>

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Label */}
        <div>
          <label className="block font-label font-medium text-ink mb-2">
            Label *
          </label>
          <input
            type="text"
            value={label}
            onChange={(e) => setLabel(e.target.value)}
            className="w-full px-4 py-2 border border-line rounded-lg bg-paper text-ink focus:outline-none focus:border-lavender-400"
            placeholder="Home, Work, etc."
          />
          {errors.label && <p className="text-rose-600 text-sm mt-1">{errors.label}</p>}
        </div>

        {/* Line 1 */}
        <div>
          <label className="block font-label font-medium text-ink mb-2">
            Address Line 1 *
          </label>
          <input
            type="text"
            value={line1}
            onChange={(e) => setLine1(e.target.value)}
            className="w-full px-4 py-2 border border-line rounded-lg bg-paper text-ink focus:outline-none focus:border-lavender-400"
            placeholder="Street address"
          />
          {errors.line1 && <p className="text-rose-600 text-sm mt-1">{errors.line1}</p>}
        </div>

        {/* Line 2 */}
        <div>
          <label className="block font-label font-medium text-ink mb-2">
            Address Line 2
          </label>
          <input
            type="text"
            value={line2}
            onChange={(e) => setLine2(e.target.value)}
            className="w-full px-4 py-2 border border-line rounded-lg bg-paper text-ink focus:outline-none focus:border-lavender-400"
            placeholder="Apartment, suite, etc. (optional)"
          />
        </div>

        {/* City */}
        <div>
          <label className="block font-label font-medium text-ink mb-2">
            City *
          </label>
          <input
            type="text"
            value={city}
            onChange={(e) => setCity(e.target.value)}
            className="w-full px-4 py-2 border border-line rounded-lg bg-paper text-ink focus:outline-none focus:border-lavender-400"
            placeholder="Accra, Kumasi, etc."
          />
          {errors.city && <p className="text-rose-600 text-sm mt-1">{errors.city}</p>}
        </div>

        {/* Region */}
        <div>
          <label className="block font-label font-medium text-ink mb-2">
            Region *
          </label>
          <input
            type="text"
            value={region}
            onChange={(e) => setRegion(e.target.value)}
            className="w-full px-4 py-2 border border-line rounded-lg bg-paper text-ink focus:outline-none focus:border-lavender-400"
            placeholder="Greater Accra, Ashanti, etc."
          />
          {errors.region && <p className="text-rose-600 text-sm mt-1">{errors.region}</p>}
        </div>

        {/* Phone */}
        <div>
          <label className="block font-label font-medium text-ink mb-2">
            Phone *
          </label>
          <input
            type="tel"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            className="w-full px-4 py-2 border border-line rounded-lg bg-paper text-ink focus:outline-none focus:border-lavender-400"
            placeholder="0201234567"
          />
          {errors.phone && <p className="text-rose-600 text-sm mt-1">{errors.phone}</p>}
        </div>

        {/* Actions */}
        <div className="flex gap-4">
          <Button
            type="submit"
            variant="primary"
            isLoading={isSubmitting}
          >
            {address ? 'Update Address' : 'Add Address'}
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
        </div>
      </form>
    </div>
  );
}
