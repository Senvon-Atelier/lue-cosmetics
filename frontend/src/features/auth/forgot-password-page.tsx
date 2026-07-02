import { useState, FormEvent } from 'react';
import { Link } from '@tanstack/react-router';
import { postAuthPasswordResetRequest } from '../../lib/api/generated/rueCosmeticsAPI';
import { Icon } from '../shared/ui/icons';
import { AuthShell } from './auth-shell';

export function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);

    try {
      await postAuthPasswordResetRequest({ email });
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to request password reset');
    } finally {
      setIsLoading(false);
    }
  };

  if (success) {
    return (
      <AuthShell
        title="Check your inbox."
        sub="Reset password"
        footer={
          <div className="auth-meta">
            Remembered? <Link to="/login">Back to sign in</Link>
          </div>
        }
      >
        <p className="auth-meta" style={{ marginTop: 0, textAlign: 'left' }}>
          We've sent a password reset link to <strong>{email}</strong>. The link will expire in 24 hours.
          If you don't see it, check your spam folder.
        </p>
      </AuthShell>
    );
  }

  return (
    <AuthShell
      title="No worries."
      sub="Reset password"
      footer={
        <div className="auth-meta">
          Remembered? <Link to="/login">Back to sign in</Link>
        </div>
      }
    >
      <form onSubmit={handleSubmit}>
        <div className="field" style={{ marginBottom: 24 }}>
          <label htmlFor="email">Email</label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            placeholder="you@somewhere.com"
          />
        </div>
        {error && (
          <div className="auth-meta" style={{ color: 'var(--lavender-700)', marginBottom: 16 }}>
            {error}
          </div>
        )}
        <button
          type="submit"
          className="btn btn-primary"
          style={{ width: '100%', justifyContent: 'center' }}
          disabled={isLoading}
        >
          Send reset link <Icon name="arrow" size={14} />
        </button>
      </form>
    </AuthShell>
  );
}
