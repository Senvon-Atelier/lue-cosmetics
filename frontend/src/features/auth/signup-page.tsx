import { useState, FormEvent } from 'react';
import { useNavigate, Link } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { Icon } from '../shared/ui/icons';
import { AuthShell } from './auth-shell';

export function SignupPage() {
  const navigate = useNavigate();
  const { signup } = useAuth();

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [agreeToTerms, setAgreeToTerms] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

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

    // Validation
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    if (!agreeToTerms) {
      setError('Please agree to the Terms of Service');
      return;
    }

    setIsLoading(true);

    try {
      await signup(email, password, name);
      // Redirect to account page or show verification message
      navigate({ to: '/account' });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create account');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthShell
      title="Begin your ritual."
      sub="Create account"
      footer={
        <div className="auth-meta">
          Already have an account? <Link to="/login">Sign in</Link>
        </div>
      }
    >
      <form onSubmit={handleSubmit}>
        <div className="field" style={{ marginBottom: 16 }}>
          <label htmlFor="name">Full Name</label>
          <input
            id="name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            placeholder="Jane Doe"
          />
        </div>
        <div className="field" style={{ marginBottom: 16 }}>
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
        <div className="field" style={{ marginBottom: 16 }}>
          <label htmlFor="password">Password</label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
            placeholder="At least 8 characters"
          />
          <span className="field-hint">Use 8+ characters with a mix of letters, numbers, and symbols.</span>
          {password && (
            <span className="field-hint">
              Password strength: <strong>{passwordStrength.strength}</strong>
            </span>
          )}
        </div>
        <div className="field" style={{ marginBottom: 16 }}>
          <label htmlFor="confirmPassword">Confirm Password</label>
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
        <label style={{ display: 'flex', gap: 10, alignItems: 'flex-start', fontFamily: 'var(--font-label)', fontSize: 12, color: 'var(--ink-soft)', marginBottom: 24 }}>
          <input
            type="checkbox"
            checked={agreeToTerms}
            onChange={(e) => setAgreeToTerms(e.target.checked)}
            required
            style={{ marginTop: 3 }}
          />
          <span>
            I agree to the{' '}
            <a href="#terms" className="auth-link">Terms of Service</a>
            {' '}and{' '}
            <a href="#privacy" className="auth-link">Privacy Policy</a>.
          </span>
        </label>
        {error && (
          <div className="auth-meta" style={{ color: 'var(--lavender-700)', marginBottom: 16 }}>
            {error}
          </div>
        )}
        <button
          type="submit"
          className="btn btn-primary"
          style={{ width: '100%', justifyContent: 'center' }}
          disabled={isLoading || !agreeToTerms || password !== confirmPassword || password.length < 8}
        >
          Create account <Icon name="arrow" size={14} />
        </button>
      </form>
    </AuthShell>
  );
}
