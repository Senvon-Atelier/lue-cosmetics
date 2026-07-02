import { useState } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { postAuthVerifyEmailResend } from '../../lib/api/generated/rueCosmeticsAPI';
import { Icon } from '../shared/ui/icons';
import { AuthShell } from './auth-shell';

export function VerifyEmailPage() {
  const navigate = useNavigate();
  const { user, refreshSession } = useAuth();

  const [isLoading, setIsLoading] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Get email from URL query params (for new signups)
  const emailParam = new URLSearchParams(window.location.search).get('email') || undefined;

  const userEmail = emailParam || user?.email;

  const handleResend = async () => {
    if (!userEmail) {
      setError('No email address found');
      return;
    }

    setError(null);
    setSuccessMessage(null);
    setIsLoading(true);

    try {
      await postAuthVerifyEmailResend();
      setSuccessMessage(`Verification email sent to ${userEmail}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to resend verification email');
    } finally {
      setIsLoading(false);
    }
  };

  const handleContinue = () => {
    // Refresh session and redirect
    refreshSession().then(() => {
      navigate({ to: '/account' });
    });
  };

  return (
    <AuthShell
      title="Check your inbox."
      sub="Verify email"
      footer={
        <div className="auth-meta">
          Already verified? <Link to="/account" onClick={(e) => { e.preventDefault(); handleContinue(); }}>Continue to your account</Link>
        </div>
      }
    >
      <p className="auth-meta" style={{ marginTop: 0, textAlign: 'left', color: 'var(--ink-soft)' }}>
        {userEmail ? (
          <>We've sent a verification email to <strong>{userEmail}</strong>.</>
        ) : (
          'Please check your email for a verification link.'
        )}
      </p>

      {successMessage && (
        <p className="auth-meta" style={{ marginTop: 0, textAlign: 'left', color: 'var(--lavender-700)', marginBottom: 16 }}>
          {successMessage}
        </p>
      )}

      {error && (
        <p className="auth-meta" style={{ marginTop: 0, textAlign: 'left', color: 'var(--lavender-700)', marginBottom: 16 }}>
          {error}
        </p>
      )}

      <div style={{ display: 'flex', flexDirection: 'column', gap: 12, marginTop: 24 }}>
        <button
          type="button"
          className="btn btn-primary"
          style={{ width: '100%', justifyContent: 'center' }}
          disabled={isLoading}
          onClick={handleResend}
        >
          Resend verification email <Icon name="arrow" size={14} />
        </button>
      </div>

      <p className="auth-meta" style={{ marginTop: 24, textAlign: 'left', color: 'var(--ink-muted)', fontSize: 12 }}>
        Didn't receive it? Check your spam folder, make sure the email address is correct, and wait a few minutes before resending.
      </p>
    </AuthShell>
  );
}
