import { useState, FormEvent, useEffect } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { postAuthPasswordResetConfirm } from '../../lib/api/generated/rueCosmeticsAPI';
import { Icon } from '../shared/ui/icons';
import { AuthShell } from './auth-shell';

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
      <AuthShell
        title="Choose a new one."
        sub="Reset password"
        footer={
          <div className="auth-meta">
            <Link to="/forgot-password">Request a new reset link</Link>
          </div>
        }
      >
        <p className="auth-meta" style={{ marginTop: 0, textAlign: 'left', color: 'var(--lavender-700)' }}>
          This password reset link is invalid or has expired.
        </p>
      </AuthShell>
    );
  }

  if (success) {
    return (
      <AuthShell
        title="Choose a new one."
        sub="Reset password"
        footer={
          <div className="auth-meta">
            <Link to="/login">Back to sign in</Link>
          </div>
        }
      >
        <p className="auth-meta" style={{ marginTop: 0, textAlign: 'left' }}>
          Your password has been successfully reset. You can now sign in with your new password.
        </p>
        <button
          type="button"
          className="btn btn-primary"
          style={{ width: '100%', justifyContent: 'center' }}
          onClick={() => navigate({ to: '/login' })}
        >
          Go to sign in <Icon name="arrow" size={14} />
        </button>
      </AuthShell>
    );
  }

  return (
    <AuthShell
      title="Choose a new one."
      sub="Reset password"
      footer={
        <div className="auth-meta">
          <Link to="/login">Back to sign in</Link>
        </div>
      }
    >
      <form onSubmit={handleSubmit}>
        <div className="field" style={{ marginBottom: 16 }}>
          <label htmlFor="password">New password</label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
            placeholder="••••••••"
          />
          {password && (
            <span className="field-hint">
              Password strength: <strong>{passwordStrength.strength}</strong>
            </span>
          )}
        </div>
        <div className="field" style={{ marginBottom: 24 }}>
          <label htmlFor="confirmPassword">Confirm new password</label>
          <input
            id="confirmPassword"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            minLength={8}
            placeholder="••••••••"
          />
          {confirmPassword && password !== confirmPassword && (
            <span className="field-error">Passwords do not match</span>
          )}
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
          disabled={isLoading || password !== confirmPassword || password.length < 8}
        >
          Save new password <Icon name="arrow" size={14} />
        </button>
      </form>
    </AuthShell>
  );
}
