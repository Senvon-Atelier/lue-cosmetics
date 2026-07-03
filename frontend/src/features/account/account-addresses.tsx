import { useEffect, useState } from 'react';
import {
  deleteMeAddressesId,
  getMeAddresses,
  patchMeAddressesId,
  postMeAddresses,
  postMeAddressesIdDefault,
} from '../../lib/api/generated/rueCosmeticsAPI';
import { GHANA_REGIONS } from '../../content/regions';
import { Icon } from '../shared/ui/icons';
import { AcctHead } from './acct-primitives';

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

type AddressFormData = {
  label: string;
  line1: string;
  line2: string;
  city: string;
  region: string;
  phone: string;
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
      <main className="acct-main">
        <div className="acct-empty">
          <p>Loading addresses…</p>
        </div>
      </main>
    );
  }

  return (
    <main className="acct-main">
      <AcctHead eyebrow="Delivery" title="Address book">
        <button className="btn btn-primary" onClick={() => { setEditingAddress(null); setShowForm(true); }}>
          <Icon name="plus" size={14} /> Add address
        </button>
      </AcctHead>

      {error && <div className="alert alert-warn">{error}</div>}

      {showForm && (
        <AddressForm
          key={editingAddress?.id ?? 'new'}
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

      {addresses.length === 0 && !showForm ? (
        <div className="acct-empty">
          <p>No addresses yet — add one to make checkout easier.</p>
          <button className="btn btn-primary" onClick={() => { setEditingAddress(null); setShowForm(true); }}>
            Add your first address
          </button>
        </div>
      ) : (
        <div className="addr-grid">
          {addresses.map((a) => (
            <div key={a.id} className={`addr-card ${a.is_default ? 'default' : ''}`}>
              {a.is_default && <span className="pill">Default</span>}
              <h4>{a.label || 'Address'}</h4>
              <p>
                {a.line1}
                {a.line2 && (
                  <>
                    <br />
                    {a.line2}
                  </>
                )}
                <br />
                {[a.city, a.region].filter(Boolean).join(', ')}
                <br />
                {a.phone}
              </p>
              <div className="actions">
                <button
                  onClick={() => {
                    setEditingAddress(a);
                    setShowForm(true);
                  }}
                >
                  Edit
                </button>
                {!a.is_default && (
                  <button onClick={() => a.id && handleSetDefault(a.id)}>
                    Set default
                  </button>
                )}
                <button className="danger" onClick={() => a.id && handleDelete(a.id)}>
                  Remove
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </main>
  );
}

function AddressForm({
  address,
  onSubmit,
  onCancel,
}: {
  address: Address | null;
  onSubmit: (data: AddressFormData) => Promise<void>;
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

    const newErrors: Record<string, string> = {};
    if (!label.trim()) newErrors.label = 'Label is required';
    if (!line1.trim()) newErrors.line1 = 'Street address is required';
    if (!city.trim()) newErrors.city = 'City is required';
    if (!region.trim()) newErrors.region = 'Region is required';
    if (!phone.trim()) newErrors.phone = 'Phone is required';

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      setIsSubmitting(false);
      return;
    }

    try {
      await onSubmit({ label, line1, line2, city, region, phone });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="form-card" style={{ marginBottom: 24, maxWidth: 'none' }}>
      <h3 className="form-card-title">
        {address ? 'Edit address' : 'New address'}
      </h3>
      <form onSubmit={handleSubmit}>
        <div className="form-row">
          <div className="field">
            <label>Label</label>
            <input
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              placeholder="e.g. Home"
            />
            {errors.label && <span className="field-error">{errors.label}</span>}
          </div>
          <div className="field">
            <label>Phone</label>
            <input
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder="+233 24 000 0000"
            />
            {errors.phone && <span className="field-error">{errors.phone}</span>}
          </div>
        </div>
        <div className="form-row full">
          <div className="field">
            <label>Street address</label>
            <input
              value={line1}
              onChange={(e) => setLine1(e.target.value)}
              placeholder="14 Amilcar Cabral Ave"
            />
            {errors.line1 && <span className="field-error">{errors.line1}</span>}
          </div>
        </div>
        <div className="form-row full">
          <div className="field">
            <label>Apartment, suite, etc. (optional)</label>
            <input value={line2} onChange={(e) => setLine2(e.target.value)} />
          </div>
        </div>
        <div className="form-row">
          <div className="field">
            <label>City</label>
            <input
              value={city}
              onChange={(e) => setCity(e.target.value)}
              placeholder="Accra"
            />
            {errors.city && <span className="field-error">{errors.city}</span>}
          </div>
          <div className="field">
            <label>Region</label>
            <select value={region} onChange={(e) => setRegion(e.target.value)}>
              <option value="">Select a region…</option>
              {GHANA_REGIONS.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
            {errors.region && <span className="field-error">{errors.region}</span>}
          </div>
        </div>
        <div className="acct-head-actions">
          <button className="btn btn-primary" type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Saving…' : address ? 'Update address' : 'Save address'}
          </button>
          <button className="btn btn-ghost" type="button" onClick={onCancel} disabled={isSubmitting}>
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
