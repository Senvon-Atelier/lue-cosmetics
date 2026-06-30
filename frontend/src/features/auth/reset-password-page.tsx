import { useState, FormEvent, useEffect } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { postAuthPasswordResetConfirm } from '../../lib/api/generated/rueCosmeticsAPI';
import { Button } from '../shared/ui/button';

export function ResetPasswordPage() {
  const navigate = useNavigate();

  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isValidToken, setIsValidToken] = useState(true);

  // Get token from URL query params
  const token = new URLSearchParams(window.location.search).get('token') || undefined;

  useEffect(() => {
    if (!token) {
      setIsValidToken(false);
      setError('Invalid or missing reset token. Please request a new password reset.');
    }
  }, [token]);

  // Password strength checker
  const getPasswordStrength = (pwd: string): { strength: string; score: number } => {
    let score = 0;
    if (pwd.length >= 8) score++;
    if (pwd.length >= 12) score++;
    if (/[a-z]/.test(pwd) && /[A-Z]/.test(pwd)) score++;
    if (/\d/.test(pwd)) score++;
    if (/[!@#$%^&*(),.?":{}|<>]/.test(pwd)) score++;

    if (score <= 2) return { strength: 'Weak', score: 1 };
    if (score === 3) return { strength: 'Fair', score: 2 };
    if (score === 4) return { strength: 'Good', score: 3 };
    return { strength: 'Strong', score: 4 };
  };

  const passwordStrength = getPasswordStrength(password);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!token) {
      setError('Invalid reset token');
      return;
    }

    // Validation
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    setIsLoading(true);

    try {
      await postAuthPasswordResetConfirm({
        token,
        new_password: password,
      });
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to reset password');
    } finally {
      setIsLoading(false);
    }
  };

  if (!isValidToken) {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center py-12 px-4">
        <div className="max-w-md w-full space-y-8 text-center">
          {/* Invalid token */}
          <div className="bg-rose-50 border border-rose-200 text-rose-800 px-6 py-8 rounded-lg">
            <div className="text-5xl mb-4">⚠️</div>
            <h1 className="font-display text-2xl mb-2">Invalid Reset Link</h1>
            <p className="text-ink-muted mb-4">
              This password reset link is invalid or has expired.
            </p>
          </div>

          {/* Request new link */}
          <div className="text-sm">
            <Link
              to="/forgot-password"
              className="text-lavender-600 hover:text-lavender-700 font-label font-medium transition-colors"
            >
              Request a new reset link
            </Link>
          </div>
        </div>
      </div>
    );
  }

  if (success) {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center py-12 px-4">
        <div className="max-w-md w-full space-y-8 text-center">
          {/* Success message */}
          <div className="bg-green-50 border border-green-200 text-green-800 px-6 py-8 rounded-lg">
            <div className="text-5xl mb-4">✓</div>
            <h1 className="font-display text-2xl mb-2">Password Reset</h1>
            <p className="text-ink-muted">
              Your password has been successfully reset.
              <br />
              You can now log in with your new password.
            </p>
          </div>

          {/* Login button */}
          <Button
            variant="primary"
            size="lg"
            onClick={() => navigate({ to: '/login' })}
            className="w-full"
          >
            Go to Login
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center py-12 px-4">
      <div className="max-w-md w-full space-y-8">
        {/* Header */}
        <div className="text-center">
          <h1 className="font-display text-4xl mb-2">Reset Password</h1>
          <p className="text-ink-muted">Enter your new password below</p>
        </div>

        {/* Error message */}
        {error && (
          <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg text-sm">
            {error}
          </div>
        )}

        {/* Reset form */}
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* New password field */}
          <div>
            <label htmlFor="password" className="block font-label font-medium text-ink mb-2">
              New Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              className="w-full px-4 py-2 border-b border-line bg-transparent text-ink placeholder:text-ink-muted focus:outline-none focus:border-lavender-400 focus:ring-1 focus:ring-lavender-400 transition-colors"
              placeholder="••••••••"
            />

            {/* Password strength indicator */}
            {password && (
              <div className="mt-2">
                <div className="flex gap-1 mb-1">
                  {[1, 2, 3, 4].map((level) => (
                    <div
                      key={level}
                      className={`h-1 flex-1 rounded-full transition-colors ${
                        passwordStrength.score >= level
                          ? level === 1
                            ? 'bg-rose-400'
                            : level === 2
                            ? 'bg-yellow-400'
                            : level === 3
                            ? 'bg-lavender-400'
                            : 'bg-green-400'
                          : 'bg-gray-200'
                      }`}
                    />
                  ))}
                </div>
                <p className="text-xs text-ink-muted">
                  Password strength: <span className="font-medium">{passwordStrength.strength}</span>
                </p>
              </div>
            )}
          </div>

          {/* Confirm password field */}
          <div>
            <label htmlFor="confirmPassword" className="block font-label font-medium text-ink mb-2">
              Confirm New Password
            </label>
            <input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
              minLength={8}
              className="w-full px-4 py-2 border-b border-line bg-transparent text-ink placeholder:text-ink-muted focus:outline-none focus:border-lavender-400 focus:ring-1 focus:ring-lavender-400 transition-colors"
              placeholder="••••••••"
            />
            {confirmPassword && password !== confirmPassword && (
              <p className="mt-1 text-xs text-rose-600">Passwords do not match</p>
            )}
          </div>

          {/* Submit button */}
          <Button
            type="submit"
            variant="primary"
            size="lg"
            isLoading={isLoading}
            disabled={password !== confirmPassword || password.length < 8}
            className="w-full"
          >
            Reset Password
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
