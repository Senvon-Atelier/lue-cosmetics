import { useState } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { postAuthVerifyEmailResend } from '../../lib/api/generated/rueCosmeticsAPI';
import { Button } from '../shared/ui/button';

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
    <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center py-12 px-4">
      <div className="max-w-md w-full space-y-8">
        {/* Header */}
        <div className="text-center">
          <div className="text-5xl mb-4">📧</div>
          <h1 className="font-display text-4xl mb-2">Verify Your Email</h1>
          <p className="text-ink-muted">
            {userEmail ? (
              <>
                We've sent a verification email to
                <br />
                <span className="font-medium text-ink">{userEmail}</span>
              </>
            ) : (
              'Please check your email for a verification link'
            )}
          </p>
        </div>

        {/* Info box */}
        <div className="bg-lavender-50 border border-lavender-200 text-ink px-4 py-3 rounded-lg text-sm">
          <p className="font-medium mb-1">Why verify?</p>
          <p className="text-ink-muted">
            Verifying your email helps us secure your account and send you important updates about your orders.
          </p>
        </div>

        {/* Success message */}
        {successMessage && (
          <div className="bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded-lg text-sm">
            {successMessage}
          </div>
        )}

        {/* Error message */}
        {error && (
          <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg text-sm">
            {error}
          </div>
        )}

        {/* Actions */}
        <div className="space-y-4">
          {/* Resend email button */}
          <Button
            variant="outline"
            size="lg"
            isLoading={isLoading}
            onClick={handleResend}
            className="w-full"
          >
            Resend Verification Email
          </Button>

          {/* Continue button */}
          <Button
            variant="primary"
            size="lg"
            onClick={handleContinue}
            className="w-full"
          >
            I've Verified My Email →
          </Button>
        </div>

        {/* Help text */}
        <div className="text-center text-sm text-ink-muted space-y-2">
          <p>Didn't receive the email?</p>
          <ul className="space-y-1">
            <li>• Check your spam folder</li>
            <li>• Make sure the email address is correct</li>
            <li>• Wait a few minutes before resending</li>
          </ul>
        </div>

        {/* Back to account/home */}
        <div className="text-center text-sm">
          <Link
            to="/"
            className="text-lavender-600 hover:text-lavender-700 font-label font-medium transition-colors"
          >
            ← Back to home
          </Link>
        </div>
      </div>
    </div>
  );
}
