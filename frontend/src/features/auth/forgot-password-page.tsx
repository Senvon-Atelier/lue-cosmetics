import { useState, FormEvent } from 'react';
import { Link } from '@tanstack/react-router';
import { postAuthPasswordResetRequest } from '../../lib/api/generated/rueCosmeticsAPI';
import { Button } from '../shared/ui/button';

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
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center py-12 px-4">
        <div className="max-w-md w-full space-y-8 text-center">
          {/* Success message */}
          <div className="bg-green-50 border border-green-200 text-green-800 px-6 py-8 rounded-lg">
            <div className="text-5xl mb-4">📧</div>
            <h1 className="font-display text-2xl mb-2">Check Your Email</h1>
            <p className="text-ink-muted mb-4">
              We've sent a password reset link to
              <br />
              <span className="font-medium text-ink">{email}</span>
            </p>
            <p className="text-sm text-ink-muted">
              The link will expire in 24 hours. If you don't see it, check your spam folder.
            </p>
          </div>

          {/* Back to login */}
          <div className="text-sm">
            <Link
              to="/login"
              className="text-lavender-600 hover:text-lavender-700 font-label font-medium transition-colors"
            >
              ← Back to login
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center py-12 px-4">
      <div className="max-w-md w-full space-y-8">
        {/* Header */}
        <div className="text-center">
          <h1 className="font-display text-4xl mb-2">Forgot Password?</h1>
          <p className="text-ink-muted">Enter your email and we'll send you a reset link</p>
        </div>

        {/* Error message */}
        {error && (
          <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg text-sm">
            {error}
          </div>
        )}

        {/* Email form */}
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Email field */}
          <div>
            <label htmlFor="email" className="block font-label font-medium text-ink mb-2">
              Email
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="w-full px-4 py-2 border-b border-line bg-transparent text-ink placeholder:text-ink-muted focus:outline-none focus:border-lavender-400 focus:ring-1 focus:ring-lavender-400 transition-colors"
              placeholder="your@email.com"
            />
          </div>

          {/* Submit button */}
          <Button
            type="submit"
            variant="primary"
            size="lg"
            isLoading={isLoading}
            className="w-full"
          >
            Send Reset Link
          </Button>
        </form>

        {/* Back to login */}
        <div className="text-center text-sm">
          <Link
            to="/login"
            className="text-lavender-600 hover:text-lavender-700 font-label font-medium transition-colors"
          >
            ← Back to login
          </Link>
        </div>
      </div>
    </div>
  );
}
